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
	"sync"
)

type PlayAction int

const (
	PlayRewind PlayAction = iota
	PlayContinue
	PlayPause
	PlayResume
	PlayStop
)

type PlayOptions struct {
	Action PlayAction
	Wait   bool
	Loop   bool
}

type soundMgr struct {
	g        *Game
	playersM sync.Mutex
	audios   map[string]Sound
}

func (p *soundMgr) init(g *Game) {
	p.audios = make(map[string]Sound)
	p.g = g
}

func (p *soundMgr) playAction(media Sound, opts *PlayOptions) (err error) {
	switch opts.Action {
	case PlayRewind:
		err = p.play(media, opts.Wait, opts.Loop)
	case PlayContinue:
		err = p.playContinue(media, opts.Wait, opts.Loop)
	case PlayStop:
		p.stop(media)
	case PlayResume:
		p.resume(media)
	case PlayPause:
		p.pause(media)
	}
	return
}

func (p *soundMgr) stopAll() {
}

func (p *soundMgr) playContinue(media Sound, wait, loop bool) (err error) {

	return
}

func (p *soundMgr) play(media Sound, wait, loop bool) (err error) {

	return
}

func (p *soundMgr) stop(media Sound) {

}

func (p *soundMgr) pause(media Sound) {

}

func (p *soundMgr) resume(media Sound) {

}

func (p *soundMgr) volume() float64 {
	return 0
}

func (p *soundMgr) SetVolume(volume float64) {

}

func (p *soundMgr) ChangeVolume(delta float64) {
}

// -------------------------------------------------------------------------------------
