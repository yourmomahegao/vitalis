package tasks

import (
	"context"
	"log"
	"time"
	"vitalis/internal/database"
	"vitalis/internal/enviroment"

	"github.com/hibiken/asynq"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

type CPUInfo struct {
	Name            string  `json:"name"`
	PhysicalCores   int     `json:"physical_cores"`
	LogicalCores    int     `json:"logical_cores"`
	Utilization     float64 `json:"utilization"`
	CurrentSpeedMHz float64 `json:"current_speed_mhz"`
	BaseSpeedMHz    float64 `json:"base_speed_mhz"`
	ProcessesAmount int     `json:"processes_amount"`
	ThreadsAmount   int     `json:"threads_amount"`
	HandlesAmount   int     `json:"handles_amount"`
	Uptime          int64   `json:"uptime"`
}

type RAMInfo struct {
	Total    uint64 `json:"total"`
	Used     uint64 `json:"used"`
	Free     uint64 `json:"free"`
	Commited uint64 `json:"commited"`
	Cached   uint64 `json:"cached"`
}

type NetInfo struct {
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
	ErrIn       uint64 `json:"err_in"`
	ErrOut      uint64 `json:"err_out"`
	Connections int    `json:"connections"`
}

type FileInfo struct {
	Path        string  `json:"path"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

type ProcessInfo struct {
	PID           int32   `json:"pid"`
	Name          string  `json:"name"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float32 `json:"memory_percent"`
	Status        string  `json:"status"`
}

func CollectCPUInfo() (*CPUInfo, error) {
	info := &CPUInfo{}

	// Name
	cpuInfos, err := cpu.Info()
	if err != nil {
		log.Printf("Error while getting CPU info in CollectCPUInfo(): %v", err)
		return nil, err
	}

	if len(cpuInfos) > 0 {
		info.Name = cpuInfos[0].ModelName
		info.BaseSpeedMHz = cpuInfos[0].Mhz
	}

	// Cores and threads
	physCount, err := cpu.Counts(false)
	if err != nil {
		log.Printf("Error while getting CPU (physical) cores in CollectCPUInfo(): %v", err)
		info.PhysicalCores = 0
	} else {
		info.PhysicalCores = physCount
	}

	logicalCount, err := cpu.Counts(true)
	if err != nil {
		log.Printf("Error while getting CPU (logical) cores in CollectCPUInfo(): %v", err)
		info.LogicalCores = 0
	} else {
		info.LogicalCores = logicalCount
	}

	// Utilization
	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		log.Printf("Error while getting CPU utilization in CollectCPUInfo(): %v", err)
		info.Utilization = 0
	}

	if len(percentages) > 0 {
		info.Utilization = percentages[0]
	}

	// Speed
	perCoreInfo, err := cpu.Info()
	if err == nil && len(perCoreInfo) > 0 {
		var sum float64
		for _, c := range perCoreInfo {
			sum += c.Mhz
		}
		info.CurrentSpeedMHz = sum / float64(len(perCoreInfo))
	}

	// Processes, threads, handles
	procs, err := process.Processes()
	if err != nil {
		log.Printf("Error while getting processes in CollectCPUInfo(): %v", err)
		info.ProcessesAmount = 0
	}

	info.ProcessesAmount = len(procs)

	var totalThreads, totalHandles int
	for _, p := range procs {
		if numThreads, err := p.NumThreads(); err == nil {
			totalThreads += int(numThreads)
		}

		if numHandles, err := p.NumFDs(); err == nil {
			totalHandles += int(numHandles)
		}
	}

	info.ThreadsAmount = totalThreads
	info.HandlesAmount = totalHandles

	// Uptime
	uptimeSeconds, err := host.Uptime()
	if err != nil {
		log.Printf("Error while getting uptime: %v", err)
	}

	info.Uptime = int64(time.Duration(uptimeSeconds) * time.Second)

	return info, nil
}

func CollectCPUInfoTask(ctx context.Context, t *asynq.Task) error {
	info, err := CollectCPUInfo()
	if err != nil {
		log.Printf("Error while getting CPU info in CollectCPUInfoTask(): %v", err)
		return err
	}

	_, err = database.Database.Exec(`delete from vitalis.public.info_cpu 
										where id in (select 
														id from (select row_number() over (order by id desc) as id_rn, 
														id 
													from vitalis.public.info_cpu) 
														where id_rn > $1)`, enviroment.Env.MaxInfoRowsAmount-1)
	if err != nil {
		log.Printf("Error while deleting old CPU info in CollectCPUInfoTask(): %v", err)
		return err
	}

	_, err = database.Database.Exec(`insert into vitalis.public.info_cpu
										(name, physical_cores, logical_cores, utilization, current_speed_mhz, base_speed_mhz, processes_amount, threads_amount, handles_amount, uptime)
									values
										($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`, info.Name, info.PhysicalCores, info.LogicalCores, info.Utilization, info.CurrentSpeedMHz, info.BaseSpeedMHz, info.ProcessesAmount, info.ThreadsAmount, info.HandlesAmount, info.Uptime)

	if err != nil {
		log.Printf("Error while inseting CPU info in CollectCPUInfoTask(): %v", err)
		return err
	}

	return nil
}

func CollectRAMInfo() (*RAMInfo, error) {
	ramInfo := &RAMInfo{}

	memoryStat, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("Error while getting virtual memory info in CollectRAMInfo(): %v", err)
		return nil, err
	}

	ramInfo.Total = memoryStat.Total
	ramInfo.Used = memoryStat.Used
	ramInfo.Free = memoryStat.Free
	ramInfo.Commited = memoryStat.CommittedAS
	ramInfo.Cached = memoryStat.Cached

	return ramInfo, nil
}

func CollectRAMInfoTask(ctx context.Context, t *asynq.Task) error {
	info, err := CollectRAMInfo()
	if err != nil {
		log.Printf("Error while getting RAM info in CollectRAMInfoTask(): %v", err)
		return err
	}

	_, err = database.Database.Exec(`delete from vitalis.public.info_ram 
										where id in (select 
														id from (select row_number() over (order by id desc) as id_rn, 
														id 
													from vitalis.public.info_ram) 
														where id_rn > $1)`, enviroment.Env.MaxInfoRowsAmount-1)
	if err != nil {
		log.Printf("Error while deleting old RAM info in CollectRAMInfoTask(): %v", err)
		return err
	}

	_, err = database.Database.Exec(`insert into vitalis.public.info_ram
										(total, used, free, commited, cached)
									values
										($1, $2, $3, $4, $5)`, info.Total, info.Used, info.Free, info.Commited, info.Cached)

	if err != nil {
		log.Printf("Error while inseting RAM info in CollectRAMInfoTask(): %v", err)
		return err
	}

	return nil
}

func CollectNetInfo() (*NetInfo, error) {
	netInfo := &NetInfo{}

	ioCounters, err := net.IOCounters(false) // false — суммарно по всем интерфейсам
	if err != nil {
		log.Printf("Error while getting net IO counters in CollectNetInfo(): %v", err)
		return nil, err
	}

	if len(ioCounters) > 0 {
		netInfo.BytesSent = ioCounters[0].BytesSent
		netInfo.BytesRecv = ioCounters[0].BytesRecv
		netInfo.PacketsSent = ioCounters[0].PacketsSent
		netInfo.PacketsRecv = ioCounters[0].PacketsRecv
		netInfo.ErrIn = ioCounters[0].Errin
		netInfo.ErrOut = ioCounters[0].Errout
	}

	connections, err := net.Connections("all")
	if err != nil {
		log.Printf("Error while getting net connections in CollectNetInfo(): %v", err)
		return nil, err
	}

	netInfo.Connections = len(connections)

	return netInfo, nil
}

func CollectNetInfoTask(ctx context.Context, t *asynq.Task) error {
	info, err := CollectNetInfo()
	if err != nil {
		log.Printf("Error while getting net info in CollectNetInfoTask(): %v", err)
		return err
	}

	log.Printf("NetInfo: %v", info)

	return nil
}

func CollectFileInfo() ([]*FileInfo, error) {
	partitions, err := disk.Partitions(false) // false — без псевдо-ФС (tmpfs и т.д.)
	if err != nil {
		log.Printf("Error while getting disk partitions in CollectFileInfo(): %v", err)
		return nil, err
	}

	fileInfos := make([]*FileInfo, 0, len(partitions))

	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			log.Printf("Error while getting disk usage for %s in CollectFileInfo(): %v", partition.Mountpoint, err)
			continue
		}

		fileInfos = append(fileInfos, &FileInfo{
			Path:        partition.Mountpoint,
			Total:       usage.Total,
			Used:        usage.Used,
			Free:        usage.Free,
			UsedPercent: usage.UsedPercent,
		})
	}

	return fileInfos, nil
}

func CollectFileInfoTask(ctx context.Context, t *asynq.Task) error {
	info, err := CollectFileInfo()
	if err != nil {
		log.Printf("Error while getting file info in CollectFileInfoTask(): %v", err)
		return err
	}

	log.Printf("FileInfo: %v", info)

	return nil
}

func CollectProccessesInfo() ([]*ProcessInfo, error) {
	procs, err := process.Processes()
	if err != nil {
		log.Printf("Error while getting processes in CollectProccessesInfo(): %v", err)
		return nil, err
	}

	processInfos := make([]*ProcessInfo, 0, len(procs))

	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}

		cpuPercent, err := p.CPUPercent()
		if err != nil {
			cpuPercent = 0
		}

		memPercent, err := p.MemoryPercent()
		if err != nil {
			memPercent = 0
		}

		statuses, err := p.Status()
		status := ""
		if err == nil && len(statuses) > 0 {
			status = statuses[0]
		}

		processInfos = append(processInfos, &ProcessInfo{
			PID:           p.Pid,
			Name:          name,
			CPUPercent:    cpuPercent,
			MemoryPercent: memPercent,
			Status:        status,
		})
	}

	return processInfos, nil
}

func CollectProccessesInfoTask(ctx context.Context, t *asynq.Task) error {
	info, err := CollectProccessesInfo()
	if err != nil {
		log.Printf("Error while getting processes info in CollectProccessesInfoTask(): %v", err)
		return err
	}

	log.Printf("ProcessesInfo count: %d", len(info))

	return nil
}

func CompileTasks() {
	// Producer
	redisOpt := asynq.RedisClientOpt{Addr: "localhost:6379"}
	client := asynq.NewClient(redisOpt)

	scheduler := asynq.NewScheduler(redisOpt, nil)

	// Pereodic tasks
	cpuCollectTask := asynq.NewTask("cpu:collect", nil)
	ramCollectTask := asynq.NewTask("ram:collect", nil)
	netCollectTask := asynq.NewTask("net:collect", nil)
	fileCollectTask := asynq.NewTask("file:collect", nil)
	procCollectTask := asynq.NewTask("proc:collect", nil)

	// scheduler.Register("@every 10s", cpuCollectTask, asynq.MaxRetry(3))
	// scheduler.Register("@every 10s", ramCollectTask, asynq.MaxRetry(3))
	// scheduler.Register("@every 10s", netCollectTask, asynq.MaxRetry(3))
	// scheduler.Register("@every 10s", fileCollectTask, asynq.MaxRetry(3))
	// scheduler.Register("@every 10s", procCollectTask, asynq.MaxRetry(3))

	// Intial tasks
	client.Enqueue(cpuCollectTask)
	client.Enqueue(ramCollectTask)
	client.Enqueue(netCollectTask)
	client.Enqueue(fileCollectTask)
	client.Enqueue(procCollectTask)

	// Worker
	srv := asynq.NewServer(asynq.RedisClientOpt{Addr: "localhost:6379"}, asynq.Config{Concurrency: 10})
	mux := asynq.NewServeMux()
	mux.HandleFunc("cpu:collect", CollectCPUInfoTask)
	mux.HandleFunc("ram:collect", CollectRAMInfoTask)
	mux.HandleFunc("net:collect", CollectNetInfoTask)
	mux.HandleFunc("file:collect", CollectFileInfoTask)
	mux.HandleFunc("proc:collect", CollectProccessesInfoTask)

	go scheduler.Run()
	go srv.Run(mux)
}
