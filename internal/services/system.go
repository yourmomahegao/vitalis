package services

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"vitalis/internal/database"
	"vitalis/internal/database/models"
)

type InfoType struct {
	Name  string
	Table string
}

type InfoFilter struct {
	Name            string
	SubQuery        string
	CompatibleTypes []InfoType
}

var InfoTypes = struct {
	CPU  InfoType
	RAM  InfoType
	Net  InfoType
	File InfoType
}{
	CPU:  InfoType{Name: "cpu", Table: "public.info_cpu ic"},
	RAM:  InfoType{Name: "ram", Table: "public.info_ram ir"},
	Net:  InfoType{Name: "net", Table: "public.info_net nt"},
	File: InfoType{Name: "file", Table: "public.info_file if"},
}

var InfoFilters = struct {
	// Global filters
	StartDatetime InfoFilter
	EndDatetime   InfoFilter
	Limit         InfoFilter
	Offset        InfoFilter

	// Cpu filters
	CPUName               InfoFilter
	CPUPhysicalCoresMin   InfoFilter
	CPUPhysicalCoresMax   InfoFilter
	CPULogicalCoresMin    InfoFilter
	CPULogicalCoresMax    InfoFilter
	CPUUtilizationMin     InfoFilter
	CPUUtilizationMax     InfoFilter
	CPUCurrentSpeedMHzMin InfoFilter
	CPUCurrentSpeedMHzMax InfoFilter
	CPUBaseSpeedMHzMin    InfoFilter
	CPUBaseSpeedMHzMax    InfoFilter
	CPUProcessesAmountMin InfoFilter
	CPUProcessesAmountMax InfoFilter
	CPUThreadsAmountMin   InfoFilter
	CPUThreadsAmountMax   InfoFilter
	CPUHandlesAmountMin   InfoFilter
	CPUHandlesAmountMax   InfoFilter
	CPUUptimeMin          InfoFilter
	CPUUptimeMax          InfoFilter

	// Ram filters
	RAMTotalMin    InfoFilter
	RAMTotalMax    InfoFilter
	RAMUsedMin     InfoFilter
	RAMUsedMax     InfoFilter
	RAMFreeMin     InfoFilter
	RAMFreeMax     InfoFilter
	RAMCommitedMin InfoFilter
	RAMCommitedMax InfoFilter
	RAMCachedMin   InfoFilter
	RAMCachedMax   InfoFilter

	// Net filters
	NetBytesSentMin   InfoFilter
	NetBytesSentMax   InfoFilter
	NetBytesRecvMin   InfoFilter
	NetBytesRecvMax   InfoFilter
	NetPacketsSentMin InfoFilter
	NetPacketsSentMax InfoFilter
	NetPacketsRecvMin InfoFilter
	NetPacketsRecvMax InfoFilter
	NetErrInMin       InfoFilter
	NetErrInMax       InfoFilter
	NetErrOutMin      InfoFilter
	NetErrOutMax      InfoFilter
	NetConnectionsMin InfoFilter
	NetConnectionsMax InfoFilter

	// File filters
	FilePath           InfoFilter
	FileTotalMin       InfoFilter
	FileTotalMax       InfoFilter
	FileUsedMin        InfoFilter
	FileUsedMax        InfoFilter
	FileFreeMin        InfoFilter
	FileFreeMax        InfoFilter
	FileUsedPercentMin InfoFilter
	FileUsedPercentMax InfoFilter
}{
	StartDatetime: InfoFilter{Name: "start_datetime", SubQuery: "insertion_datetime >= $0", CompatibleTypes: []InfoType{InfoTypes.CPU, InfoTypes.RAM, InfoTypes.Net, InfoTypes.File}},
	EndDatetime:   InfoFilter{Name: "end_datetime", SubQuery: "insertion_datetime <= $0", CompatibleTypes: []InfoType{InfoTypes.CPU, InfoTypes.RAM, InfoTypes.Net, InfoTypes.File}},
	Limit:         InfoFilter{Name: "limit", SubQuery: "limit $0", CompatibleTypes: []InfoType{InfoTypes.CPU, InfoTypes.RAM, InfoTypes.Net, InfoTypes.File}},
	Offset:        InfoFilter{Name: "offset", SubQuery: "offset $0", CompatibleTypes: []InfoType{InfoTypes.CPU, InfoTypes.RAM, InfoTypes.Net, InfoTypes.File}},

	CPUName:               InfoFilter{Name: "cpu_name", SubQuery: "ic.name like $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	CPUPhysicalCoresMin:   InfoFilter{Name: "cpu_physical_cores_min", SubQuery: "ic.physical_cores >= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	CPUPhysicalCoresMax:   InfoFilter{Name: "cpu_physical_cores_max", SubQuery: "ic.physical_cores <= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	CPULogicalCoresMin:    InfoFilter{Name: "cpu_logical_cores_min", SubQuery: "ic.logical_cores >= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	CPULogicalCoresMax:    InfoFilter{Name: "cpu_logical_cores_max", SubQuery: "ic.logical_cores <= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	CPUUtilizationMin:     InfoFilter{Name: "cpu_utilization_min", SubQuery: "ic.utilization >= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	CPUUtilizationMax:     InfoFilter{Name: "cpu_utilization_max", SubQuery: "ic.utilization <= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	CPUCurrentSpeedMHzMin: InfoFilter{Name: "cpu_current_speed_mhz_min", SubQuery: "ic.current_speed_mhz >= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	CPUCurrentSpeedMHzMax: InfoFilter{Name: "cpu_current_speed_mhz_max", SubQuery: "ic.current_speed_mhz <= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	CPUBaseSpeedMHzMin:    InfoFilter{Name: "cpu_base_speed_mhz_min", SubQuery: "ic.base_speed_mhz >= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	CPUBaseSpeedMHzMax:    InfoFilter{Name: "cpu_base_speed_mhz_max", SubQuery: "ic.base_speed_mhz <= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	CPUProcessesAmountMin: InfoFilter{Name: "cpu_processes_amount_min", SubQuery: "ic.processes_amount >= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	CPUProcessesAmountMax: InfoFilter{Name: "cpu_processes_amount_max", SubQuery: "ic.processes_amount <= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	CPUThreadsAmountMin:   InfoFilter{Name: "cpu_threads_amount_min", SubQuery: "ic.threads_amount >= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	CPUThreadsAmountMax:   InfoFilter{Name: "cpu_threads_amount_max", SubQuery: "ic.threads_amount <= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	CPUHandlesAmountMin:   InfoFilter{Name: "cpu_handles_amount_min", SubQuery: "ic.handles_amount >= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	CPUHandlesAmountMax:   InfoFilter{Name: "cpu_handles_amount_max", SubQuery: "ic.handles_amount <= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	CPUUptimeMin:          InfoFilter{Name: "cpu_uptime_min", SubQuery: "ic.uptime >= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	CPUUptimeMax:          InfoFilter{Name: "cpu_uptime_max", SubQuery: "ic.uptime <= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},

	RAMTotalMin:    InfoFilter{Name: "ram_total_min", SubQuery: "ir.total >= $0", CompatibleTypes: []InfoType{InfoTypes.RAM}},
	RAMTotalMax:    InfoFilter{Name: "ram_total_max", SubQuery: "ir.total <= $0", CompatibleTypes: []InfoType{InfoTypes.RAM}},
	RAMUsedMin:     InfoFilter{Name: "ram_used_min", SubQuery: "ir.used >= $0", CompatibleTypes: []InfoType{InfoTypes.RAM}},
	RAMUsedMax:     InfoFilter{Name: "ram_used_max", SubQuery: "ir.used <= $0", CompatibleTypes: []InfoType{InfoTypes.RAM}},
	RAMFreeMin:     InfoFilter{Name: "ram_free_min", SubQuery: "ir.free >= $0", CompatibleTypes: []InfoType{InfoTypes.RAM}},
	RAMFreeMax:     InfoFilter{Name: "ram_free_max", SubQuery: "ir.free <= $0", CompatibleTypes: []InfoType{InfoTypes.RAM}},
	RAMCommitedMin: InfoFilter{Name: "ram_commited_min", SubQuery: "ir.commited >= $0", CompatibleTypes: []InfoType{InfoTypes.RAM}},
	RAMCommitedMax: InfoFilter{Name: "ram_commited_max", SubQuery: "ir.commited <= $0", CompatibleTypes: []InfoType{InfoTypes.RAM}},
	RAMCachedMin:   InfoFilter{Name: "ram_cached_min", SubQuery: "ir.cached >= $0", CompatibleTypes: []InfoType{InfoTypes.RAM}},
	RAMCachedMax:   InfoFilter{Name: "ram_cached_max", SubQuery: "ir.cached <= $0", CompatibleTypes: []InfoType{InfoTypes.RAM}},

	NetBytesSentMin:   InfoFilter{Name: "net_bytes_sent_min", SubQuery: "nt.bytes_sent >= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},
	NetBytesSentMax:   InfoFilter{Name: "net_bytes_sent_max", SubQuery: "nt.bytes_sent <= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},
	NetBytesRecvMin:   InfoFilter{Name: "net_bytes_recv_min", SubQuery: "nt.bytes_recv >= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},
	NetBytesRecvMax:   InfoFilter{Name: "net_bytes_recv_max", SubQuery: "nt.bytes_recv <= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},
	NetPacketsSentMin: InfoFilter{Name: "net_packets_sent_min", SubQuery: "nt.packets_sent >= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},
	NetPacketsSentMax: InfoFilter{Name: "net_packets_sent_max", SubQuery: "nt.packets_sent <= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},
	NetPacketsRecvMin: InfoFilter{Name: "net_packets_recv_min", SubQuery: "nt.packets_recv >= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},
	NetPacketsRecvMax: InfoFilter{Name: "net_packets_recv_max", SubQuery: "nt.packets_recv <= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},
	NetErrInMin:       InfoFilter{Name: "net_err_in_min", SubQuery: "nt.err_in >= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},
	NetErrInMax:       InfoFilter{Name: "net_err_in_max", SubQuery: "nt.err_in <= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},
	NetErrOutMin:      InfoFilter{Name: "net_err_out_min", SubQuery: "nt.err_out >= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},
	NetErrOutMax:      InfoFilter{Name: "net_err_out_max", SubQuery: "nt.err_out <= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},
	NetConnectionsMin: InfoFilter{Name: "net_connections_min", SubQuery: "nt.connections >= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},
	NetConnectionsMax: InfoFilter{Name: "net_connections_max", SubQuery: "nt.connections <= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},

	FilePath:           InfoFilter{Name: "file_path", SubQuery: "if.path like $0", CompatibleTypes: []InfoType{InfoTypes.File}},
	FileTotalMin:       InfoFilter{Name: "file_total_min", SubQuery: "if.total >= $0", CompatibleTypes: []InfoType{InfoTypes.File}},
	FileTotalMax:       InfoFilter{Name: "file_total_max", SubQuery: "if.total <= $0", CompatibleTypes: []InfoType{InfoTypes.File}},
	FileUsedMin:        InfoFilter{Name: "file_used_min", SubQuery: "if.used >= $0", CompatibleTypes: []InfoType{InfoTypes.File}},
	FileUsedMax:        InfoFilter{Name: "file_used_max", SubQuery: "if.used <= $0", CompatibleTypes: []InfoType{InfoTypes.File}},
	FileFreeMin:        InfoFilter{Name: "file_free_min", SubQuery: "if.free >= $0", CompatibleTypes: []InfoType{InfoTypes.File}},
	FileFreeMax:        InfoFilter{Name: "file_free_max", SubQuery: "if.free <= $0", CompatibleTypes: []InfoType{InfoTypes.File}},
	FileUsedPercentMin: InfoFilter{Name: "file_used_percent_min", SubQuery: "if.used_percent >= $0", CompatibleTypes: []InfoType{InfoTypes.File}},
	FileUsedPercentMax: InfoFilter{Name: "file_used_percent_max", SubQuery: "if.used_percent <= $0", CompatibleTypes: []InfoType{InfoTypes.File}},
}

func checkFiltersCompability(infoType InfoType, infoFilters []InfoFilter) (bool, error) {
	if len(infoFilters) == 0 {
		return true, nil
	}

	for _, infoFilter := range infoFilters {
		for _, compatibleInfoType := range infoFilter.CompatibleTypes {
			if compatibleInfoType == infoType {
				return true, nil
			}
		}
	}

	return false, fmt.Errorf("None of filters compatible with type %s", infoType.Name)
}

func compileFilters(query string, infoFilters []InfoFilter) (string, error) {
	var whereClauses []string
	var limitClause, offsetClause string

	placeholderIndex := 1

	for _, infoFilter := range infoFilters {
		subQuery := strings.Replace(infoFilter.SubQuery, "$0", fmt.Sprintf("$%d", placeholderIndex), 1)
		placeholderIndex++

		switch infoFilter.Name {
		case InfoFilters.Limit.Name:
			limitClause = subQuery
		case InfoFilters.Offset.Name:
			offsetClause = subQuery
		default:
			whereClauses = append(whereClauses, subQuery)
		}
	}

	if len(whereClauses) > 0 {
		query += " where " + strings.Join(whereClauses, " and ")
	}

	if limitClause != "" {
		query += " " + limitClause
	}

	if offsetClause != "" {
		query += " " + offsetClause
	}

	return query, nil
}

func compileRows(infoType InfoType, rows *sql.Rows) (any, error) {
	defer rows.Close()

	switch infoType.Name {
	case InfoTypes.CPU.Name:
		result := []models.InfoCPU{}

		for rows.Next() {
			row := models.InfoCPU{}

			err := rows.Scan(&row.Id, &row.GroupId, &row.Name, &row.PhysicalCores, &row.LogicalCores, &row.Utilization, &row.CurrentSpeedMHz, &row.BaseSpeedMHz, &row.ProcessesAmount, &row.ThreadsAmount, &row.HandlesAmount, &row.Uptime, &row.InsertionDatetime)
			if err != nil {
				log.Printf("Error occured while scanning CPU row in compileRows(): %v", err)
				return nil, err
			}

			result = append(result, row)
		}

		return result, rows.Err()
	case InfoTypes.RAM.Name:
		result := []models.InfoRAM{}

		for rows.Next() {
			row := models.InfoRAM{}

			err := rows.Scan(&row.Id, &row.GroupId, &row.Total, &row.Used, &row.Free, &row.Commited, &row.Cached, &row.InsertionDatetime)
			if err != nil {
				log.Printf("Error occured while scanning RAM row in compileRows(): %v", err)
				return nil, err
			}

			result = append(result, row)
		}

		return result, rows.Err()
	case InfoTypes.Net.Name:
		result := []models.InfoNet{}

		for rows.Next() {
			row := models.InfoNet{}

			err := rows.Scan(&row.Id, &row.GroupId, &row.BytesSent, &row.BytesRecv, &row.PacketsSent, &row.PacketsRecv, &row.ErrIn, &row.ErrOut, &row.Connections, &row.InsertionDatetime)
			if err != nil {
				log.Printf("Error occured while scanning net row in compileRows(): %v", err)
				return nil, err
			}

			result = append(result, row)
		}

		return result, rows.Err()
	case InfoTypes.File.Name:
		result := []models.InfoFile{}

		for rows.Next() {
			row := models.InfoFile{}

			err := rows.Scan(&row.Id, &row.GroupId, &row.Path, &row.Total, &row.Used, &row.Free, &row.UsedPercent, &row.InsertionDatetime)
			if err != nil {
				log.Printf("Error occured while scanning file row in compileRows(): %v", err)
				return nil, err
			}

			result = append(result, row)
		}

		return result, rows.Err()
	default:
		return nil, fmt.Errorf("Unknown info type: %s", infoType.Name)
	}
}

func GetInfoData(infoType InfoType, infoFilters []InfoFilter, filterValues []any) (any, error) {
	_, compError := checkFiltersCompability(infoType, infoFilters)

	if compError != nil {
		log.Printf("Error occured while checking filters compability in GetInfoData(): %v", compError)
		return nil, compError
	}

	rawQuery := fmt.Sprintf(`select
								*
							from %s`, infoType.Table)

	query, err := compileFilters(rawQuery, infoFilters)

	if err != nil {
		log.Printf("Error occured while compiling filters in GetInfoData(): %v", err)
		return nil, err
	}

	rows, queryErr := database.Database.Query(query, filterValues...)

	if queryErr != nil {
		log.Printf("Error occured while processing query in GetInfoData(): %v", queryErr)
		return nil, queryErr
	}

	result, err := compileRows(infoType, rows)
	if err != nil {
		log.Printf("Error occured while compiling rows in GetInfoData(): %v", err)
		return nil, err
	}

	return result, nil
}
