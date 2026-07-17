package tasks

import (
	"context"
	"fmt"
	"log"
	"time"
	"vitalis/internal/database"
	"vitalis/internal/enviroment"
	"vitalis/internal/tasks/structs"

	"github.com/hibiken/asynq"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

func CollectCPUInfo() (*structs.CPUInfo, error) {
	info := &structs.CPUInfo{}

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

	var groupId int = 0
	groupIdSeqRow := database.Database.QueryRow(`select nextval('info_cpu_group_id_seq');`)

	err = groupIdSeqRow.Scan(&groupId)
	if err != nil {
		log.Printf("Error while getting new group_id in CollectCPUInfoTask(): %v", err)
		return err
	}

	_, err = database.Database.Exec(`delete from public.info_cpu
										where group_id in (
											select group_id from (
												select group_id,
													dense_rank() over (order by group_id desc) as group_rnk
												from public.info_cpu
											) t
											where group_rnk > $1
										);`, enviroment.ENV.MAX_INFO_GROUPS_AMOUNT-1)
	if err != nil {
		log.Printf("Error while deleting old CPU info in CollectCPUInfoTask(): %v", err)
		return err
	}

	_, err = database.Database.Exec(`insert into public.info_cpu
										(group_id, name, physical_cores, logical_cores, utilization, current_speed_mhz, base_speed_mhz, processes_amount, threads_amount, handles_amount, uptime)
									values
										($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`, groupId, info.Name, info.PhysicalCores, info.LogicalCores, info.Utilization, info.CurrentSpeedMHz, info.BaseSpeedMHz, info.ProcessesAmount, info.ThreadsAmount, info.HandlesAmount, info.Uptime)

	if err != nil {
		log.Printf("Error while inseting CPU info in CollectCPUInfoTask(): %v", err)
		return err
	}

	return nil
}

func CollectRAMInfo() (*structs.RAMInfo, error) {
	ramInfo := &structs.RAMInfo{}

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

	var groupId int = 0
	groupIdSeqRow := database.Database.QueryRow(`select nextval('info_ram_group_id_seq');`)

	err = groupIdSeqRow.Scan(&groupId)
	if err != nil {
		log.Printf("Error while getting new group_id in CollectRAMInfoTask(): %v", err)
		return err
	}

	_, err = database.Database.Exec(`delete from public.info_ram
										where group_id in (
											select group_id from (
												select group_id,
													dense_rank() over (order by group_id desc) as group_rnk
												from public.info_ram
											) t
											where group_rnk > $1
										);`, enviroment.ENV.MAX_INFO_GROUPS_AMOUNT-1)
	if err != nil {
		log.Printf("Error while deleting old RAM info in CollectRAMInfoTask(): %v", err)
		return err
	}

	_, err = database.Database.Exec(`insert into public.info_ram
										(group_id, total, used, free, commited, cached)
									values
										($1, $2, $3, $4, $5, $6)`, groupId, info.Total, info.Used, info.Free, info.Commited, info.Cached)

	if err != nil {
		log.Printf("Error while inseting RAM info in CollectRAMInfoTask(): %v", err)
		return err
	}

	return nil
}

func CollectNetInfo() (*structs.NetInfo, error) {
	netInfo := &structs.NetInfo{}

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

	var groupId int = 0
	groupIdSeqRow := database.Database.QueryRow(`select nextval('info_net_group_id_seq');`)

	err = groupIdSeqRow.Scan(&groupId)
	if err != nil {
		log.Printf("Error while getting new group_id in CollectNetInfoTask(): %v", err)
		return err
	}

	_, err = database.Database.Exec(`delete from public.info_net
										where group_id in (
											select group_id from (
												select group_id,
													dense_rank() over (order by group_id desc) as group_rnk
												from public.info_net
											) t
											where group_rnk > $1
										);`, enviroment.ENV.MAX_INFO_GROUPS_AMOUNT-1)
	if err != nil {
		log.Printf("Error while deleting old NET info in CollectNetInfoTask(): %v", err)
		return err
	}

	_, err = database.Database.Exec(`insert into public.info_net
										(group_id, bytes_sent, bytes_recv, packets_sent, packets_recv, err_in, err_out, connections)
									values
										($1, $2, $3, $4, $5, $6, $7, $8)`, groupId, info.BytesSent, info.BytesRecv, info.PacketsSent, info.PacketsRecv, info.ErrIn, info.ErrOut, info.Connections)

	if err != nil {
		log.Printf("Error while inseting NET info in CollectNetInfoTask(): %v", err)
		return err
	}

	return nil
}

func CollectFileInfo() ([]*structs.FileInfo, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		log.Printf("Error while getting disk partitions in CollectFileInfo(): %v", err)
		return nil, err
	}

	fileInfos := make([]*structs.FileInfo, 0, len(partitions))

	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			log.Printf("Error while getting disk usage for %s in CollectFileInfo(): %v", partition.Mountpoint, err)
			continue
		}

		fileInfos = append(fileInfos, &structs.FileInfo{
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

	var groupId int = 0
	groupIdSeqRow := database.Database.QueryRow(`select nextval('info_file_group_id_seq');`)

	err = groupIdSeqRow.Scan(&groupId)
	if err != nil {
		log.Printf("Error while getting new group_id in CollectFileInfoTask(): %v", err)
		return err
	}

	_, err = database.Database.Exec(`delete from public.info_file
										where group_id in (
											select group_id from (
												select group_id,
													dense_rank() over (order by group_id desc) as group_rnk
												from public.info_file
											) t
											where group_rnk > $1
										);`, enviroment.ENV.MAX_INFO_GROUPS_AMOUNT-1)
	if err != nil {
		log.Printf("Error while deleting old files info in CollectFileInfoTask(): %v", err)
		return err
	}

	for _, infoRow := range info {
		_, err = database.Database.Exec(`insert into public.info_file
											(group_id, path, total, used, free, used_percent)
										values
											($1, $2, $3, $4, $5, $6)`, groupId, infoRow.Path, infoRow.Total, infoRow.Used, infoRow.Free, infoRow.UsedPercent)
	}

	if err != nil {
		log.Printf("Error while inseting files info in CollectFileInfoTask(): %v", err)
		return err
	}

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

	scheduler.Register(fmt.Sprintf("@every %ds", enviroment.ENV.COLLECT_CPU_INFO_INTERVAL_SECONDS), cpuCollectTask, asynq.MaxRetry(3))
	scheduler.Register(fmt.Sprintf("@every %ds", enviroment.ENV.COLLECT_RAM_INFO_INTERVAL_SECONDS), ramCollectTask, asynq.MaxRetry(3))
	scheduler.Register(fmt.Sprintf("@every %ds", enviroment.ENV.COLLECT_NET_INFO_INTERVAL_SECONDS), netCollectTask, asynq.MaxRetry(3))
	scheduler.Register(fmt.Sprintf("@every %ds", enviroment.ENV.COLLECT_FILE_INFO_INTERVAL_SECONDS), fileCollectTask, asynq.MaxRetry(3))

	// Intial tasks
	client.Enqueue(cpuCollectTask)
	client.Enqueue(ramCollectTask)
	client.Enqueue(netCollectTask)
	client.Enqueue(fileCollectTask)

	// Worker
	srv := asynq.NewServer(asynq.RedisClientOpt{Addr: "localhost:6379"}, asynq.Config{Concurrency: 10})
	mux := asynq.NewServeMux()
	mux.HandleFunc("cpu:collect", CollectCPUInfoTask)
	mux.HandleFunc("ram:collect", CollectRAMInfoTask)
	mux.HandleFunc("net:collect", CollectNetInfoTask)
	mux.HandleFunc("file:collect", CollectFileInfoTask)

	go scheduler.Run()
	go srv.Run(mux)
}
