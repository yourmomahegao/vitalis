package structs

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
