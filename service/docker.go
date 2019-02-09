package service

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	influx "github.com/influxdata/influxdb1-client"
	"golang.org/x/net/context"
	resty "gopkg.in/resty.v1"
)

var (
	DockerSocket     = "/var/run/docker.sock"
	DockerAPIVersion = "v1.39"
	RunningStats     = map[string]bool{}
)

func RunDockerStats(ctx context.Context, containerID string) error {
	logging(ctx).Debugf("Connecting to docker: %s", DockerSocket)
	if _, ok := RunningStats[containerID]; ok {
		logging(ctx).Debugf("Already running...")
		return nil
	}

	// Create a Go's http.Transport so we can set it in resty.
	transport := http.Transport{
		Dial: func(_, _ string) (net.Conn, error) {
			return net.Dial("unix", DockerSocket)
		},
	}

	// Set the previous transport that we created, set the scheme of the communication to the
	// socket and set the unixSocket as the HostURL.
	client := resty.New().SetTransport(&transport).SetScheme("http").SetHostURL(DockerSocket)
	statURL := fmt.Sprintf("http://%s/containers/%s/stats", DockerAPIVersion, containerID)
	logging(ctx).Debugf("GET %s", statURL)

	db, err := influxClient()
	if err != nil {
		return err
	}

	influxSetRetentionPolicy(db)
	// ch := make(chan struct{})
	// reqCtx, cancel := context.WithCancel(ctx)
	// go func() {
	// 	<-ch // wait for server to start request handling
	// 	cancel()
	// }()

	// client.SetDoNotParseResponse(true)
	req := client.R()
	req.SetDoNotParseResponse(true)
	req.SetContext(ctx)
	resp, err := req.Get(statURL)
	//curl --unix-socket /var/run/docker.sock http://v1.37/containers/5bee4c18ad99/stats
	//https://docs.docker.com/engine/api/v1.30/#operation/ContainerStats

	if err != nil {
		return err
	}
	// explore response object
	// logging(ctx).Debugf("Response Status Code: %v", resp.StatusCode())
	// logging(ctx).Debugf("Response Status: %v", resp.Status())
	// logging(ctx).Debugf("Response Time: %v", resp.Time())
	// logging(ctx).Debugf("Response Received At: %v", resp.ReceivedAt())
	// logging(ctx).Debugf("Response Body: %v", resp) // or resp.String() or string(resp.Body())

	reader := bufio.NewReader(resp.RawResponse.Body)
	go func() error {
		pts := []influx.Point{}

		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				logging(ctx).Debugf("ended? %+v", err)
				break
			}
			// logging(ctx).Debugf("Body: %s", line)
			var data map[string]interface{}
			if err := json.Unmarshal(line, &data); err != nil {
				return err
			}

			// logging(ctx).Debugf("date: %s, pre: %s", data["read"], data["preread"])
			if data["networks"] == nil {
				return fmt.Errorf("Unable to get stats")
			}
			networks := data["networks"].(map[string]interface{})
			readBandwidth := float64(0)
			writeBandwidth := float64(0)
			for _, rawValues := range networks {
				values := rawValues.(map[string]interface{})
				// logging(ctx).Debugf("network: %s bandwidth: %f", name, values["tx_bytes"])
				readBandwidth += values["rx_bytes"].(float64)
				writeBandwidth += values["tx_bytes"].(float64)
			}

			memories := data["memory_stats"].(map[string]interface{})["stats"].(map[string]interface{})
			memory := float64(0)
			memory = memories["rss"].(float64)

			// todo based on calculation https://stackoverflow.com/questions/30271942/get-docker-container-cpu-usage-as-percentage
			cpuStats := data["cpu_stats"].(map[string]interface{})["cpu_usage"].(map[string]interface{})
			cpu := cpuStats["total_usage"].(float64)

			cpuStats = data["cpu_stats"].(map[string]interface{})
			systemCPU := cpuStats["system_cpu_usage"].(float64)

			// preCPUStats := data["precpu_stats"].(map[string]interface{})["cpu_usage"].(map[string]interface{})
			// preCPU := preCPUStats["total_usage"].(float64)

			// preCPUStats = data["cpu_stats"].(map[string]interface{})
			// preSystemCPU := preCPUStats["system_cpu_usage"].(float64)

			// logging(ctx).Debugf("cpu: %f, system %f, preSystemCPU %f", cpu, systemCPU, preSystemCPU)
			// cpu = cpu - preCPU
			// systemCPU = systemCPU - preSystemCPU
			cpuPercent := cpu / systemCPU * 100

			readStats := float64(0)
			writeStats := float64(0)
			storedStats := data["blkio_stats"].(map[string]interface{})["io_service_bytes_recursive"].([]interface{})
			for _, rawStore := range storedStats {
				store := rawStore.(map[string]interface{})
				if store["op"] == "Read" {
					readStats = store["value"].(float64)
				} else if store["op"] == "Write" {
					writeStats = store["value"].(float64)
				}
			}

			logging(ctx).Debugf("%s memory: %fMB, cpu: %f, io r/w: %f/%f, network: r/w: %fMB/%fMB", containerID, memory/1024/1024, cpuPercent, readStats, writeStats, readBandwidth/1024/1024, writeBandwidth/1024/1024)
			// logging(ctx).Debugf("Body data ...")

			pts = append(pts,
				influx.Point{
					Measurement: "containers",
					Tags:        map[string]string{},
					Fields: map[string]interface{}{
						"name":        containerID,
						"cpu_raw":     cpu,
						"cpu_system":  systemCPU,
						"cpu_percent": cpuPercent,
						"io_read":     readStats,
						"io_write":    writeStats,
						"memory":      memory,
						"net_read":    readBandwidth,
						"net_write":   writeBandwidth,
					},
					// Time:      time.Unix(0, 0).UTC(),
					Precision: "s",
				})

			bps := influx.BatchPoints{
				Points:          pts,
				Database:        "report",
				RetentionPolicy: "6months",
			}
			_, err = db.Write(bps)
			if err != nil {
				log.Fatal(err)
			}

			// logging(ctx).Debugf("Saved")
			pts = []influx.Point{}
		} //each stats

		return nil
	}()

	RunningStats[containerID] = true
	// select {
	// case <-ctx.Done():
	// 	logging(ctx).Debugf("Cancelling server?")
	// }

	return nil
}

/*
{
  "read":"2019-01-20T07:50:09.5991469Z",
  "preread":"2019-01-20T07:50:08.5953927Z",
  "pids_stats":{
    "current":17
  },
  "blkio_stats":{
    "io_service_bytes_recursive":[
      {
        "major":8,
        "minor":0,
        "op":"Read",
        "value":0
      },
      {
        "major":8,
        "minor":0,
        "op":"Write",
        "value":8192
      },
      {
        "major":8,
        "minor":0,
        "op":"Sync",
        "value":8192
      },
      {
        "major":8,
        "minor":0,
        "op":"Async",
        "value":0
      },
      {
        "major":8,
        "minor":0,
        "op":"Total",
        "value":8192
      }
    ],
    "io_serviced_recursive":[
      {
        "major":8,
        "minor":0,
        "op":"Read",
        "value":0
      },
      {
        "major":8,
        "minor":0,
        "op":"Write",
        "value":58
      },
      {
        "major":8,
        "minor":0,
        "op":"Sync",
        "value":58
      },
      {
        "major":8,
        "minor":0,
        "op":"Async",
        "value":0
      },
      {
        "major":8,
        "minor":0,
        "op":"Total",
        "value":58
      }
    ],
    "io_queue_recursive":[

    ],
    "io_service_time_recursive":[

    ],
    "io_wait_time_recursive":[

    ],
    "io_merged_recursive":[

    ],
    "io_time_recursive":[

    ],
    "sectors_recursive":[

    ]
  },
  "num_procs":0,
  "storage_stats":{

  },
  "cpu_stats":{
    "cpu_usage":{
      "total_usage":100306639300,
      "percpu_usage":[
        32722435100,
        31945237500,
        35638966700
      ],
      "usage_in_kernelmode":10950000000,
      "usage_in_usermode":3460000000
    },
    "system_cpu_usage":46796950000000,
    "online_cpus":3,
    "throttling_data":{
      "periods":0,
      "throttled_periods":0,
      "throttled_time":0
    }
  },
  "precpu_stats":{
    "cpu_usage":{
      "total_usage":100232108400,
      "percpu_usage":[
        32679376200,
        31945237500,
        35607494700
      ],
      "usage_in_kernelmode":10940000000,
      "usage_in_usermode":3450000000
    },
    "system_cpu_usage":46794120000000,
    "online_cpus":3,
    "throttling_data":{
      "periods":0,
      "throttled_periods":0,
      "throttled_time":0
    }
  },
  "memory_stats":{
    "usage":16773120,
    "max_usage":16908288,
    "stats":{
      "active_anon":15884288,
      "active_file":4096,
      "cache":28672,
      "dirty":0,
      "hierarchical_memory_limit":9223372036854771712,
      "hierarchical_memsw_limit":9223372036854771712,
      "inactive_anon":0,
      "inactive_file":24576,
      "mapped_file":0,
      "pgfault":5377,
      "pgmajfault":0,
      "pgpgin":4458,
      "pgpgout":573,
      "rss":15884288,
      "rss_huge":0,
      "total_active_anon":15884288,
      "total_active_file":4096,
      "total_cache":28672,
      "total_dirty":0,
      "total_inactive_anon":0,
      "total_inactive_file":24576,
      "total_mapped_file":0,
      "total_pgfault":5377,
      "total_pgmajfault":0,
      "total_pgpgin":4458,
      "total_pgpgout":573,
      "total_rss":15884288,
      "total_rss_huge":0,
      "total_unevictable":0,
      "total_writeback":0,
      "unevictable":0,
      "writeback":0
    },
    "limit":3149832192
  },
  "name":"/dgapidev_dnode1_1",
  "id":"a0b1036b6dbdf3e1a02b8af9c810bbeaae8d6ac62c3784b658cc135945faf1c8",
  "networks":{
    "eth0":{
      "rx_bytes":2019579,
      "rx_packets":15262,
      "rx_errors":0,
      "rx_dropped":0,
      "tx_bytes":1603588,
      "tx_packets":17032,
      "tx_errors":0,
      "tx_dropped":0
    }
  }
}
*/
