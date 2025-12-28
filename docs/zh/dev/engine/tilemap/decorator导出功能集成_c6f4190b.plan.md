---
name: Decorator导出功能集成
overview: 将 tscn_parser 中的 decorator 导出功能用 GDScript 重新实现，集成到 spx_tilemap_exporter 插件中，通过直接访问 Godot 节点（而非解析 tscn 文件）来简化处理流程。
todos:
  - id: create-decorator-extractor
    content: 创建 decorator_extractor.gd 核心类，实现节点扫描和数据提取
    status: in_progress
  - id: implement-coord-transform
    content: 实现坐标系转换逻辑（Y轴翻转、旋转、Pivot计算）
    status: pending
    dependencies:
      - create-decorator-extractor
  - id: implement-collider-extract
    content: 实现碰撞体数据提取（支持多种Shape2D类型）
    status: pending
    dependencies:
      - create-decorator-extractor
  - id: update-plugin
    content: 修改 spx_tilemap_exporter.gd 添加 Decorator 导出菜单
    status: pending
    dependencies:
      - create-decorator-extractor
  - id: update-cli
    content: 修改 export_cli.gd 支持命令行导出 decorator
    status: pending
    dependencies:
      - create-decorator-extractor
---

# Decorator 导出功能集成到 spx_tilemap_exporter

## 一、tscn_parser 导出 Decorator 的机制分析

### 1.1 数据流程

```mermaid
flowchart LR
    subgraph Godot[Godot场景]
        A[Sprite2D节点]
        B[预制体实例]
    end
    
    subgraph Parser[tscn_parser Go解析]
        C[解析tscn文本]
        D[提取DecoratorNode]
        E[解析PrefabInfo]
        F[ConvertToTilemap合并]
    end
    
    subgraph Output[输出]
        G[JSON文件]
        H[decorators数组]
    end
    
    A --> C
    B --> C
    C --> D
    C --> E
    D --> F
    E --> F
    F --> G
    G --> H
```

### 1.2 Decorator 数据结构

tscn_parser 导出的 Decorator JSON 格式（[types.go](tscn_parser/types.go) 89-103行）：

```json
{
  "name": "节点名称",
  "path": "textures/sprite.png",
  "position": {"x": 100, "y": -200},
  "scale": {"x": 1, "y": 1},
  "rotation": 90,
  "z_index": 0,
  "pivot": {"x": 16, "y": -8},
  "collider_type": "rect",
  "collider_pivot": {"x": 0, "y": 0},
  "collider_params": [32, 32]
}
```

### 1.3 坐标系转换规则

从 [converter.go](tscn_parser/converter.go) 分析的转换规则：

| 转换项 | 规则 | 原因 |

|--------|------|------|

| Position.Y | `-position.y` | Godot Y向下 → SPX Y向上 |

| Rotation | `rad * 180 / PI + 90` | 弧度转角度 + 方向补偿 |

| Pivot.Y | `-pivot.y` | Y轴翻转 |

| Pivot | `pivot + texture_size / 2` | 从左上角原点转为中心点基准 |

| ColliderPivot.Y | `-collider_pivot.y` | Y轴翻转 |

| ColliderPivot | `collider_pivot - pivot` | 转为相对于精灵锚点 |

## 二、Decorator 与 TileMap 的关系

- **TileMap**: 规则网格瓦片，按 tile 坐标排列，用于地形/背景
- **Decorator**: 自由放置的精灵，任意像素坐标，用于装饰物/可交互对象

两者在 SPX 中独立加载（[tilemap.go](tilemap.go) 122-124行）：

```go
p.loadTilemaps(p.datas)
p.loadDecorators(p.datas)
```

## 三、GDScript 实现方案

### 3.1 新增文件

在 `addons/spx_tilemap_exporter/` 目录下创建：

- `decorator_extractor.gd` - 核心提取类

### 3.2 节点扫描策略

```mermaid
flowchart TD
    A[场景根节点] --> B{遍历所有子节点}
    B --> C{是Sprite2D?}
    C -->|是| D[提取为简单Decorator]
    C -->|否| E{是预制体实例?}
    E -->|是| F[解析预制体获取碰撞信息]
    E -->|否| G[继续遍历子节点]
    D --> H[应用坐标转换]
    F --> H
    H --> I[添加到decorators数组]
```

### 3.3 碰撞体提取

支持的碰撞体类型及参数：

- `rect`: ColliderParams = [width, height]
- `circle`: ColliderParams = [radius]
- `polygon`: ColliderParams = [x1, y1, x2, y2, ...]
- `capsule`: ColliderParams = [radius, height]

### 3.4 修改现有文件

修改 [spx_tilemap_exporter.gd](tutorial/AITown/addons/spx_tilemap_exporter/spx_tilemap_exporter.gd):

- 添加 Decorator 导出菜单项
- 集成 DecoratorExtractor

修改 [export_cli.gd](tutorial/AITown/addons/spx_tilemap_exporter/export_cli.gd):

- 支持同时导出 tilemap 和 decorator

## 四、实现任务