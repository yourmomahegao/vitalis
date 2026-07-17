package models

type AuthSessionKeys struct {
	Id               int
	SessionKey       string
	CreationDatetime string
	ValidUntil       string
}

type InfoCPU struct {
	Id                int     `json:"id"`
	GroupId           int     `json:"group_id"`
	Name              string  `json:"name"`
	PhysicalCores     int     `json:"physical_cores"`
	LogicalCores      int     `json:"logical_cores"`
	Utilization       float64 `json:"utilization"`
	CurrentSpeedMHz   float64 `json:"current_speed_mhz"`
	BaseSpeedMHz      float64 `json:"base_speed_mhz"`
	ProcessesAmount   int     `json:"processes_amount"`
	ThreadsAmount     int     `json:"threads_amount"`
	HandlesAmount     int     `json:"handles_amount"`
	Uptime            int64   `json:"uptime"`
	InsertionDatetime string  `json:"insertion_datetime"`
}

type InfoRAM struct {
	Id                int    `json:"id"`
	GroupId           int    `json:"group_id"`
	Total             int64  `json:"total"`
	Used              int64  `json:"used"`
	Free              int64  `json:"free"`
	Commited          int64  `json:"commited"`
	Cached            int64  `json:"cached"`
	InsertionDatetime string `json:"insertion_datetime"`
}

type InfoNet struct {
	Id                int    `json:"id"`
	GroupId           int    `json:"group_id"`
	BytesSent         int64  `json:"bytes_sent"`
	BytesRecv         int64  `json:"bytes_recv"`
	PacketsSent       int64  `json:"packets_sent"`
	PacketsRecv       int64  `json:"packets_recv"`
	ErrIn             int64  `json:"err_in"`
	ErrOut            int64  `json:"err_out"`
	Connections       int    `json:"connections"`
	InsertionDatetime string `json:"insertion_datetime"`
}

type InfoFile struct {
	Id                int     `json:"id"`
	GroupId           int     `json:"group_id"`
	Path              string  `json:"path"`
	Total             int64   `json:"total"`
	Used              int64   `json:"used"`
	Free              int64   `json:"free"`
	UsedPercent       float64 `json:"used_percent"`
	InsertionDatetime string  `json:"insertion_datetime"`
}
