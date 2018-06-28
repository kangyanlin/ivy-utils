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
	"sync/atomic"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh/terminal"
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
	NoTTY  bool
	frames []rune
	pos    int
	active uint64
}

// Start activates the spinner
func (sp *Spinner) Start() *Spinner {
	if atomic.LoadUint64(&sp.active) > 0 {
		return sp
	}
	atomic.StoreUint64(&sp.active, 1)
	go func() {
		for atomic.LoadUint64(&sp.active) > 0 {
			fmt.Fprintf(sp.Writer, "\r\033[K%s%s%s", sp.Prefix, sp.next(), sp.Suffix)
			time.Sleep(100 * time.Millisecond)
		}
	}()
	return sp
}

// SetCharset sets custom spinner character set
func (sp *Spinner) SetCharset(frames string) {
	sp.frames = []rune(frames)
}

// Stop stops and clear-up the spinner
func (sp *Spinner) Stop() bool {
	if x := atomic.SwapUint64(&sp.active, 0); x > 0 {
		fmt.Fprintf(sp.Writer, "\r\033[K")
		return true
	}
	return false
}

func (sp *Spinner) next() string {
	r := sp.frames[sp.pos%len(sp.frames)]
	sp.pos++
	return string(r)
}

// NewSpinner creates a new spinner instance
func NewSpinner() *Spinner {
	s := &Spinner{
		Writer: os.Stderr,
	}
	s.SetCharset(Default)
	if !terminal.IsTerminal(syscall.Stderr) {
		s.NoTTY = true
	}
	return s
}
