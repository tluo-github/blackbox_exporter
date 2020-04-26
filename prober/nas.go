// Copyright 2016 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prober

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/blackbox_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	"os"
	"syscall"
)

type DiskStatus struct {
	All  uint64 `json:"all"`
	Used uint64 `json:"used"`
	Free uint64 `json:"free"`
}

// disk usage of path/disk
func DiskUsage(path string) (DiskStatus, error) {
	fs := syscall.Statfs_t{}
	disk := DiskStatus{
		All:  0,
		Used: 0,
		Free: 0,
	}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return disk, err
	}
	disk.All = fs.Blocks * uint64(fs.Bsize)
	disk.Free = fs.Bfree * uint64(fs.Bsize)
	disk.Used = disk.All - disk.Free
	return disk, nil
}

func ProbeNAS(ctx context.Context, target string, module config.Module, registry *prometheus.Registry, logger log.Logger) bool {
	diskAllGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "probe_nas_all_size_bytes",
		Help: "NAS DISK all capacity",
	})
	diskUsedGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "probe_nas_used_size_bytes",
		Help: "NAS DISK used capacity",
	})
	diskFreeGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "probe_nas_free_size_bytes",
		Help: "NAS DISK free capacity",
	})
	probeFailedTouchFile := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "probe_nas_failed_touch_file",
		Help: "Test touch file",
	})
	registry.MustRegister(diskAllGauge)
	registry.MustRegister(diskUsedGauge)
	registry.MustRegister(diskFreeGauge)
	registry.MustRegister(probeFailedTouchFile)
	disk, err := DiskUsage(target)
	if err != nil {
		level.Error(logger).Log("msg", "Error DiskUsage", "err", err)
		return false
	}
	tempFile, err := ioutil.TempFile(target, "ProbeNas_")
	if err != nil {
		probeFailedTouchFile.Set(1)
		level.Error(logger).Log("msg", "faild touch file")
		return false
	}
	defer tempFile.Close()
	defer os.Remove(tempFile.Name())
	diskAllGauge.Set(float64(disk.All))
	diskUsedGauge.Set(float64(disk.Used))
	diskFreeGauge.Set(float64(disk.Free))
	probeFailedTouchFile.Set(0)
	return true
}
