// Copyright © 2018 Alfred Chou <unioverlord@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// Spinner types.
var (
	Box1    = `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏`
	Box2    = `⠋⠙⠚⠞⠖⠦⠴⠲⠳⠓`
	Box3    = `⠄⠆⠇⠋⠙⠸⠰⠠⠰⠸⠙⠋⠇⠆`
	Box4    = `⠋⠙⠚⠒⠂⠂⠒⠲⠴⠦⠖⠒⠐⠐⠒⠓⠋`
	Box5    = `⠁⠉⠙⠚⠒⠂⠂⠒⠲⠴⠤⠄⠄⠤⠴⠲⠒⠂⠂⠒⠚⠙⠉⠁`
	Box6    = `⠈⠉⠋⠓⠒⠐⠐⠒⠖⠦⠤⠠⠠⠤⠦⠖⠒⠐⠐⠒⠓⠋⠉⠈`
	Box7    = `⠁⠁⠉⠙⠚⠒⠂⠂⠒⠲⠴⠤⠄⠄⠤⠠⠠⠤⠦⠖⠒⠐⠐⠒⠓⠋⠉⠈⠈`
	Spin1   = `|/-\`
	Spin2   = `◴◷◶◵`
	Spin3   = `◰◳◲◱`
	Spin4   = `◐◓◑◒`
	Spin5   = `▉▊▋▌▍▎▏▎▍▌▋▊▉`
	Spin6   = `▌▄▐▀`
	Spin7   = `╫╪`
	Spin8   = `■□▪▫`
	Spin9   = `←↑→↓`
	Spin10  = `⦾⦿`
	Default = Spin1
)

type Spinner struct {
	Prefix string
	Suffix string
	Writer io.Writer
	mu     sync.Mutex
	frames []rune
	length int
	pos    int
	active uint32
}

// Start activates the spinner
func (sp *Spinner) Start() *Spinner {
	if atomic.LoadUint32(&sp.active) > 0 {
		return sp
	}
	atomic.StoreUint32(&sp.active, 1)
	go func() {
		for atomic.LoadUint32(&sp.active) > 0 {
			fmt.Fprintf(sp.Writer, "\r\033[K%s%s%s", sp.Prefix, sp.Next(), sp.Suffix)
			time.Sleep(333 * time.Millisecond)
		}
	}()
	return sp
}

// SetCharset sets custom spinner character set
func (sp *Spinner) SetCharset(frames string) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.frames = []rune(frames)
	sp.length = len(sp.frames)
}

// Stop stops and clear-up the spinner
func (sp *Spinner) Stop() bool {
	if x := atomic.SwapUint32(&sp.active, 0); x > 0 {
		fmt.Fprintf(sp.Writer, "\r\033[K")
		return true
	}
	return false
}

// Current returns the current rune in the sequence.
func (sp *Spinner) Current() string {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	r := sp.frames[sp.pos%sp.length]
	return string(r)
}

// Next returns the next rune in the sequence.
func (sp *Spinner) Next() string {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	r := sp.frames[sp.pos%sp.length]
	sp.pos++
	return string(r)
}

// Reset the spinner to its initial frame.
func (sp *Spinner) Reset() {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.pos = 0
}

func (sp *Spinner) run() {

}

// NewSpinner creates a new spinner instance
func NewSpinner() *Spinner {
	s := &Spinner{
		Writer: os.Stderr,
	}
	s.SetCharset(Default)
	return s
}
