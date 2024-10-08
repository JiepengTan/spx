/*
 * Copyright (c) 2021 The GoPlus Authors (goplus.org). All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package spx

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	measureEdgeLen    = 6
	measureLineWidth  = 2
	measureTextMargin = 8
)

type measure struct {
	size    float64
	x       float64
	y       float64
	heading float64

	// computed properties
	text         string
	color        Color
	svgLineStyle string
	svgRotate    string
	svgSize      int // size*scale + 0.5 + measureLineWidth
}

func newMeasure(v specsp) *measure {
	size := v["size"].(float64)
	scale := getSpcspVal(v, "scale", 1.0).(float64)
	text := strconv.FormatFloat(size, 'f', 1, 64)
	text = strings.TrimSuffix(text, ".0")
	heading := getSpcspVal(v, "heading", 0.0).(float64)
	svgSize := int(size*scale + 0.5 + measureLineWidth)
	c, err := parseColor(getSpcspVal(v, "color", 0.0))
	if err != nil {
		panic(err)
	}
	return &measure{
		heading:      heading,
		size:         size,
		text:         text,
		color:        c,
		x:            v["x"].(float64),
		y:            v["y"].(float64),
		svgLineStyle: fmt.Sprintf("stroke-width:%d;stroke:rgb(%d, %d, %d);", measureLineWidth, c.R, c.G, c.B),
		svgRotate:    fmt.Sprintf("rotate(%.1f %d %d)", heading, svgSize>>1, svgSize>>1),
		svgSize:      svgSize,
	}
}

func getSpcspVal(ss specsp, key string, defaultVal ...interface{}) interface{} {
	v, ok := ss[key]
	if ok {
		return v
	}
	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return v
}
