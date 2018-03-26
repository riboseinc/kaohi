
/*
 * Copyright (c) 2017, [Ribose Inc](https://www.ribose.com).
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions
 * are met:
 * 1. Redistributions of source code must retain the above copyright
 *    notice, this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright
 *    notice, this list of conditions and the following disclaimer in the
 *    documentation and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
 * ``AS IS'' AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
 * LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
 * A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
 * OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
 * SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
 * LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
 * DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
 * THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// An Op is a type that is used to describe what type
// of event has occurred during the watching process.
type Op uint32

// Ops
const (
	Create Op = iota
	Write
	Remove
	Rename
	Chmod
	Move
)

var ops = map[Op]string{
	Create: "CREATE",
	Write:  "WRITE",
	Remove: "REMOVE",
	Rename: "RENAME",
	Chmod:  "CHMOD",
	Move:   "MOVE",
}

// global variable for watcher
var kWatcher *Watcher

// String prints the string version of the Op consts
func (e Op) String() string {
	if op, found := ops[e]; found {
		return op
	}
	return "UNKNOWN"
}

// An Event describes an event that is received when files or directory
// changes occur. It includes the os.FileInfo of the changed file or
// directory and the type of event that's occurred and the full path of the file.
type Event struct {
	Op
	Path string
	os.FileInfo
}

// String returns a string depending on what type of event occurred and the
// file name associated with the event.
func (e Event) String() string {
	if e.FileInfo != nil {
		pathType := "FILE"
		if e.IsDir() {
			pathType = "DIRECTORY"
		}
		return fmt.Sprintf("%s %q %s [%s]", pathType, e.Name(), e.Op, e.Path)
	}
	return "???"
}

type Watcher struct {
	Event  chan Event
	Error  chan error
	Closed chan struct{}
	close  chan struct{}
	wg     *sync.WaitGroup

	// mu protects the following.
	mu           *sync.Mutex
	files        map[string]os.FileInfo // map of files.
	ops          map[Op]struct{}        // Op filtering.
	maxEvents    int                    // max sent events per cycle
}

// New creates a new Watcher.
func NewWatcher() *Watcher {
	// Set up the WaitGroup for w.Wait().
	var wg sync.WaitGroup
	wg.Add(1)

	return &Watcher{
		Event:   make(chan Event),
		Error:   make(chan error),
		Closed:  make(chan struct{}),
		close:   make(chan struct{}),
		mu:      new(sync.Mutex),
		wg:      &wg,
		files:   make(map[string]os.FileInfo),
	}
}

// SetMaxEvents controls the maximum amount of events that are sent on
// the Event channel per watching cycle. If max events is less than 1, there is
// no limit, which is the default.
func (w *Watcher) SetMaxEvents(delta int) {
	w.mu.Lock()
	w.maxEvents = delta
	w.mu.Unlock()
}

// FilterOps filters which event op types should be returned
// when an event occurs.
func (w *Watcher) FilterOps(ops ...Op) {
	w.mu.Lock()
	w.ops = make(map[Op]struct{})
	for _, op := range ops {
		w.ops[op] = struct{}{}
	}
	w.mu.Unlock()
}

// Add adds either a single file or directory to the file list.
func (w *Watcher) AddFile(name string) (err error) {
	DEBUG_INFO("Adding file '%s' to watcher", name)

	w.mu.Lock()
	defer w.mu.Unlock()

	name, err = filepath.Abs(name)
	if err != nil {
		return err
	}

	// add file to file lists
	stat, err := os.Stat(name)
	if err != nil {
		return err
	}

	if !stat.Mode().IsRegular() {
		return err
	}
	w.files[name] = stat

	return nil
}

// Remove removes either a single file from the file's list.
func (w *Watcher) RemoveFile(name string) (err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	name, err = filepath.Abs(name)
	if err != nil {
		return err
	}

	// If name is a single file, remove it and return.
	_, found := w.files[name]
	if !found {
		return nil // Doesn't exist, just return.
	}

	delete(w.files, name)
	return nil
}

// list files
func (w *Watcher) GetInfo(name string) (os.FileInfo, error) {
	// Make sure name exists
	stat, err := os.Stat(name)
	if err != nil {
		return nil, err
	}

	return stat, nil
}

// fileInfo is an implementation of os.FileInfo that can be used
// as a mocked os.FileInfo when triggering an event when the specified
// os.FileInfo is nil.
type fileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	sys     interface{}
	dir     bool
}

func (fs *fileInfo) IsDir() bool {
	return fs.dir
}
func (fs *fileInfo) ModTime() time.Time {
	return fs.modTime
}
func (fs *fileInfo) Mode() os.FileMode {
	return fs.mode
}
func (fs *fileInfo) Name() string {
	return fs.name
}
func (fs *fileInfo) Size() int64 {
	return fs.size
}
func (fs *fileInfo) Sys() interface{} {
	return fs.sys
}

// TriggerEvent is a method that can be used to trigger an event, separate to
// the file watching process.
func (w *Watcher) TriggerEvent(eventType Op, file os.FileInfo) {
	w.Wait()
	if file == nil {
		file = &fileInfo{name: "triggered event", modTime: time.Now()}
	}
	w.Event <- Event{Op: eventType, Path: "-", FileInfo: file}
}

func (w *Watcher) retrieveFileList() map[string]os.FileInfo {
	w.mu.Lock()
	defer w.mu.Unlock()

	fileList := make(map[string]os.FileInfo)
	for k, _ := range w.files {
		fileList[k], _ = w.GetInfo(k)
	}

	return fileList
}

// Start begins the polling cycle which repeats every specified
// duration until Close is called.
func (w *Watcher) Start(d time.Duration) error {

	DEBUG_INFO("Starging watcher polling process")

	// Unblock w.Wait().
	w.wg.Done()

	for {
		// done lets the inner polling cycle loop know when the
		// current cycle's method has finished executing.
		done := make(chan struct{})

		// Any events that are found are first piped to evt before
		// being sent to the main Event channel.
		evt := make(chan Event)

		// Retrieve the file list for all watched files
		fileList := w.retrieveFileList()

		// cancel can be used to cancel the current event polling function.
		cancel := make(chan struct{})

		// Look for events.
		go func() {
			w.pollEvents(fileList, evt, cancel)
			done <- struct{}{}
		}()

		// numEvents holds the number of events for the current cycle.
		numEvents := 0

	inner:
		for {
			select {
			case <-w.close:
				close(cancel)
				close(w.Closed)
				return nil

			case event := <-evt:
				if len(w.ops) > 0 { // Filter Ops.
					_, found := w.ops[event.Op]
					if !found {
						continue
					}
				}
				numEvents++
				if w.maxEvents > 0 && numEvents > w.maxEvents {
					close(cancel)
					break inner
				}
				w.Event <- event

			case <-done: // Current cycle is finished.
				break inner
			}
		}

		// Update the file's list.
		w.mu.Lock()
		w.files = fileList
		w.mu.Unlock()

		// Sleep and then continue to the next loop iteration.
		time.Sleep(d)
	}
}

func (w *Watcher) pollEvents(files map[string]os.FileInfo, evt chan Event,
	cancel chan struct{}) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Store create and remove events for use to check for rename events.
	creates := make(map[string]os.FileInfo)
	removes := make(map[string]os.FileInfo)

	// Check for removed files.
	for path, info := range w.files {
		if _, found := files[path]; !found {
			removes[path] = info
		}
	}

	// Check for created files, writes and chmods.
	for path, info := range files {
		oldInfo, found := w.files[path]
		if !found {
			// A file was created.
			creates[path] = info
			continue
		}

		if oldInfo.ModTime() != info.ModTime() {
			select {
			case <-cancel:
				return
			case evt <- Event{Write, path, info}:
				DEBUG_INFO("WRITE event has detected for file '%s'", path)
				break
			}
		}

		if oldInfo.Mode() != info.Mode() {
			select {
			case <-cancel:
				return
			case evt <- Event{Chmod, path, info}:
				DEBUG_INFO("CHMOD event has detected for file '%s'", path)
				break
			}
		}
	}

	// Send all the remaining create and remove events.
	for path, info := range creates {
		select {
		case <-cancel:
			return
		case evt <- Event{Create, path, info}:
			DEBUG_INFO("CREATE event has detected for file '%s'", path)
			break
		}
	}
	for path, info := range removes {
		select {
		case <-cancel:
			return
		case evt <- Event{Remove, path, info}:
			DEBUG_INFO("REMOVE event has detected for file '%s'", path)
			break
		}
	}
}

// Wait blocks until the watcher is started.
func (w *Watcher) Wait() {
	w.wg.Wait()
}

func (w *Watcher) Close() {
	DEBUG_INFO("Closing watcher polling thread")

	w.mu.Lock()
	w.files = make(map[string]os.FileInfo)
	w.mu.Unlock()

	// Send a close signal to the Start method.
	w.close <- struct{}{}
}

// init watcher
func InitKaohiWatcher() error {

	DEBUG_INFO("Initializing Kaohi Watcher")

	// create new watcher	
	kWatcher = NewWatcher()

	// start watcher proc
	go kWatcher.Start(50 * time.Millisecond)

	return nil
}

// finalize watcher
func FinalizeKaohiWatcher() {

	DEBUG_INFO("Finalizing Kaohi Watcher")

	// close watcher
	kWatcher.Close()
}
