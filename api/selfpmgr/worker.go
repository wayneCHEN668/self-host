/*
Copyright (C) 2021 The Self-host Authors.
This file is part of Self-host <https://github.com/self-host/self-host>.

Self-host is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Self-host is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with Self-host.  If not, see <http://www.gnu.org/licenses/>.
*/

package selfpmgr

import (
	"time"
)

type Worker struct {
	Id        string
	URI       string
	Languages []string

	load     uint64
	timeout  time.Duration
	lastSeen time.Time
}

func (w *Worker) SetLoad(l uint64) {
	w.load = l
}

func (w *Worker) GetLoad() uint64 {
	return w.load
}

func (w *Worker) Alive() bool {
	return time.Now().Before(w.lastSeen.Add(w.timeout))
}

func NewWorker(id, uri string, langs []string, timeout time.Duration) *Worker {
	return &Worker{
		Id:        id,
		URI:       uri,
		Languages: langs,
		timeout:   timeout,
		lastSeen:  time.Now(),
	}
}
