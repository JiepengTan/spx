// Code generated by gop (Go+); DO NOT EDIT.

package main

import (
	"github.com/goplus/spx"
)

const _ = true

type Bullet struct {
	spx.Sprite
	*Game
}
type MyAircraft struct {
	spx.Sprite
	*Game
}
type Game struct {
	spx.Game
	MyAircraft MyAircraft
	Bullet     Bullet
	scores     int
	life       int
}

func (this *Game) MainEntry() {
}
func (this *Game) Main() {
	spx.Gopt_Game_Main(this, new(Bullet), new(MyAircraft))
}
//line tutorial/09-ui/Bullet.spx:1
func (this *Bullet) Main() {
//line tutorial/09-ui/Bullet.spx:1:1
	this.OnCloned__1(func() {
//line tutorial/09-ui/Bullet.spx:2:1
		this.SetXYpos(this.MyAircraft.Xpos(), this.MyAircraft.Ypos()+5)
//line tutorial/09-ui/Bullet.spx:3:1
		this.Show()
//line tutorial/09-ui/Bullet.spx:4:1
		for {
			spx.Sched()
//line tutorial/09-ui/Bullet.spx:5:1
			this.Wait(0.04)
//line tutorial/09-ui/Bullet.spx:6:1
			this.Step__0(10)
//line tutorial/09-ui/Bullet.spx:7:1
			if this.Touching(spx.Edge) {
//line tutorial/09-ui/Bullet.spx:8:1
				this.Destroy()
			}
		}
	})
}
func (this *Bullet) Classfname() string {
	return "Bullet"
}
//line tutorial/09-ui/MyAircraft.spx:2
func (this *MyAircraft) Main() {
//line tutorial/09-ui/MyAircraft.spx:2:1
	this.OnStart(func() {
//line tutorial/09-ui/MyAircraft.spx:3:1
		isIncres := true
//line tutorial/09-ui/MyAircraft.spx:4:1
		sizeStep := 0.2
//line tutorial/09-ui/MyAircraft.spx:5:1
		monotorSize := 1.0
//line tutorial/09-ui/MyAircraft.spx:6:1
		for {
			spx.Sched()
//line tutorial/09-ui/MyAircraft.spx:7:1
			stepSize := 10
//line tutorial/09-ui/MyAircraft.spx:8:1
			for
//line tutorial/09-ui/MyAircraft.spx:8:1
			i := 0; i < 10;
//line tutorial/09-ui/MyAircraft.spx:8:1
			i++ {
				spx.Sched()
//line tutorial/09-ui/MyAircraft.spx:9:1
				this.Wait(0.15)
//line tutorial/09-ui/MyAircraft.spx:10:1
				this.life += stepSize
//line tutorial/09-ui/MyAircraft.spx:11:1
				stepSize *= 10
			}
//line tutorial/09-ui/MyAircraft.spx:13:1
			monitor := spx.Gopt_Game_Gopx_GetWidget[spx.Monitor](this, "life")
//line tutorial/09-ui/MyAircraft.spx:14:1
			if isIncres {
//line tutorial/09-ui/MyAircraft.spx:15:1
				monotorSize += sizeStep
//line tutorial/09-ui/MyAircraft.spx:16:1
				if monotorSize > 2 {
//line tutorial/09-ui/MyAircraft.spx:17:1
					isIncres = false
				}
			} else {
//line tutorial/09-ui/MyAircraft.spx:20:1
				monotorSize -= sizeStep
//line tutorial/09-ui/MyAircraft.spx:21:1
				if monotorSize < 1 {
//line tutorial/09-ui/MyAircraft.spx:22:1
					isIncres = true
				}
			}
//line tutorial/09-ui/MyAircraft.spx:25:1
			monitor.SetSize(float64(monotorSize))
//line tutorial/09-ui/MyAircraft.spx:26:1
			this.life = 0
		}
	})
}
func (this *MyAircraft) Classfname() string {
	return "MyAircraft"
}
func main() {
	new(Game).Main()
}
