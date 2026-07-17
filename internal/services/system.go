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

var InfoFilters = map[string]InfoFilter{
	"start_datetime": {Name: "start_datetime", SubQuery: "insertion_datetime >= $0", CompatibleTypes: []InfoType{InfoTypes.CPU, InfoTypes.RAM, InfoTypes.Net, InfoTypes.File}},
	"end_datetime":   {Name: "end_datetime", SubQuery: "insertion_datetime <= $0", CompatibleTypes: []InfoType{InfoTypes.CPU, InfoTypes.RAM, InfoTypes.Net, InfoTypes.File}},
	"limit":          {Name: "limit", SubQuery: "limit $0", CompatibleTypes: []InfoType{InfoTypes.CPU, InfoTypes.RAM, InfoTypes.Net, InfoTypes.File}},
	"offset":         {Name: "offset", SubQuery: "offset $0", CompatibleTypes: []InfoType{InfoTypes.CPU, InfoTypes.RAM, InfoTypes.Net, InfoTypes.File}},

	"cpu_name":                  {Name: "name", SubQuery: "ic.name like $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	"cpu_physical_cores_min":    {Name: "physical_cores_min", SubQuery: "ic.physical_cores >= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	"cpu_physical_cores_max":    {Name: "physical_cores_max", SubQuery: "ic.physical_cores <= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	"cpu_logical_cores_min":     {Name: "logical_cores_min", SubQuery: "ic.logical_cores >= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	"cpu_logical_cores_max":     {Name: "logical_cores_max", SubQuery: "ic.logical_cores <= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	"cpu_utilization_min":       {Name: "utilization_min", SubQuery: "ic.utilization >= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	"cpu_utilization_max":       {Name: "utilization_max", SubQuery: "ic.utilization <= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	"cpu_current_speed_mhz_min": {Name: "current_speed_mhz_min", SubQuery: "ic.current_speed_mhz >= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	"cpu_current_speed_mhz_max": {Name: "current_speed_mhz_max", SubQuery: "ic.current_speed_mhz <= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	"cpu_base_speed_mhz_min":    {Name: "base_speed_mhz_min", SubQuery: "ic.base_speed_mhz >= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	"cpu_base_speed_mhz_max":    {Name: "base_speed_mhz_max", SubQuery: "ic.base_speed_mhz <= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	"cpu_processes_amount_min":  {Name: "processes_amount_min", SubQuery: "ic.processes_amount >= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	"cpu_processes_amount_max":  {Name: "processes_amount_max", SubQuery: "ic.processes_amount <= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	"cpu_threads_amount_min":    {Name: "threads_amount_min", SubQuery: "ic.threads_amount >= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	"cpu_threads_amount_max":    {Name: "threads_amount_max", SubQuery: "ic.threads_amount <= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	"cpu_handles_amount_min":    {Name: "handles_amount_min", SubQuery: "ic.handles_amount >= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	"cpu_handles_amount_max":    {Name: "handles_amount_max", SubQuery: "ic.handles_amount <= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	"cpu_uptime_min":            {Name: "uptime_min", SubQuery: "ic.uptime >= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},
	"cpu_uptime_max":            {Name: "uptime_max", SubQuery: "ic.uptime <= $0", CompatibleTypes: []InfoType{InfoTypes.CPU}},

	"ram_total_min":    {Name: "total_min", SubQuery: "ir.total >= $0", CompatibleTypes: []InfoType{InfoTypes.RAM}},
	"ram_total_max":    {Name: "total_max", SubQuery: "ir.total <= $0", CompatibleTypes: []InfoType{InfoTypes.RAM}},
	"ram_used_min":     {Name: "used_min", SubQuery: "ir.used >= $0", CompatibleTypes: []InfoType{InfoTypes.RAM}},
	"ram_used_max":     {Name: "used_max", SubQuery: "ir.used <= $0", CompatibleTypes: []InfoType{InfoTypes.RAM}},
	"ram_free_min":     {Name: "free_min", SubQuery: "ir.free >= $0", CompatibleTypes: []InfoType{InfoTypes.RAM}},
	"ram_free_max":     {Name: "free_max", SubQuery: "ir.free <= $0", CompatibleTypes: []InfoType{InfoTypes.RAM}},
	"ram_commited_min": {Name: "commited_min", SubQuery: "ir.commited >= $0", CompatibleTypes: []InfoType{InfoTypes.RAM}},
	"ram_commited_max": {Name: "commited_max", SubQuery: "ir.commited <= $0", CompatibleTypes: []InfoType{InfoTypes.RAM}},
	"ram_cached_min":   {Name: "cached_min", SubQuery: "ir.cached >= $0", CompatibleTypes: []InfoType{InfoTypes.RAM}},
	"ram_cached_max":   {Name: "cached_max", SubQuery: "ir.cached <= $0", CompatibleTypes: []InfoType{InfoTypes.RAM}},

	"net_bytes_sent_min":   {Name: "bytes_sent_min", SubQuery: "nt.bytes_sent >= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},
	"net_bytes_sent_max":   {Name: "bytes_sent_max", SubQuery: "nt.bytes_sent <= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},
	"net_bytes_recv_min":   {Name: "bytes_recv_min", SubQuery: "nt.bytes_recv >= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},
	"net_bytes_recv_max":   {Name: "bytes_recv_max", SubQuery: "nt.bytes_recv <= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},
	"net_packets_sent_min": {Name: "packets_sent_min", SubQuery: "nt.packets_sent >= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},
	"net_packets_sent_max": {Name: "packets_sent_max", SubQuery: "nt.packets_sent <= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},
	"net_packets_recv_min": {Name: "packets_recv_min", SubQuery: "nt.packets_recv >= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},
	"net_packets_recv_max": {Name: "packets_recv_max", SubQuery: "nt.packets_recv <= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},
	"net_err_in_min":       {Name: "err_in_min", SubQuery: "nt.err_in >= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},
	"net_err_in_max":       {Name: "err_in_max", SubQuery: "nt.err_in <= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},
	"net_err_out_min":      {Name: "err_out_min", SubQuery: "nt.err_out >= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},
	"net_err_out_max":      {Name: "err_out_max", SubQuery: "nt.err_out <= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},
	"net_connections_min":  {Name: "connections_min", SubQuery: "nt.connections >= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},
	"net_connections_max":  {Name: "connections_max", SubQuery: "nt.connections <= $0", CompatibleTypes: []InfoType{InfoTypes.Net}},

	"file_path":             {Name: "path", SubQuery: "if.path like $0", CompatibleTypes: []InfoType{InfoTypes.File}},
	"file_total_min":        {Name: "total_min", SubQuery: "if.total >= $0", CompatibleTypes: []InfoType{InfoTypes.File}},
	"file_total_max":        {Name: "total_max", SubQuery: "if.total <= $0", CompatibleTypes: []InfoType{InfoTypes.File}},
	"file_used_min":         {Name: "used_min", SubQuery: "if.used >= $0", CompatibleTypes: []InfoType{InfoTypes.File}},
	"file_used_max":         {Name: "used_max", SubQuery: "if.used <= $0", CompatibleTypes: []InfoType{InfoTypes.File}},
	"file_free_min":         {Name: "free_min", SubQuery: "if.free >= $0", CompatibleTypes: []InfoType{InfoTypes.File}},
	"file_free_max":         {Name: "free_max", SubQuery: "if.free <= $0", CompatibleTypes: []InfoType{InfoTypes.File}},
	"file_used_percent_min": {Name: "used_percent_min", SubQuery: "if.used_percent >= $0", CompatibleTypes: []InfoType{InfoTypes.File}},
	"file_used_percent_max": {Name: "used_percent_max", SubQuery: "if.used_percent <= $0", CompatibleTypes: []InfoType{InfoTypes.File}},
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
		case "limit":
			limitClause = subQuery
		case "offset":
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

	log.Println(query)

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
