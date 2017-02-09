package proxy

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/golang/glog"
	"github.com/olekukonko/tablewriter"
)

// Parse query's results and formate it,then will be print
// in command line such as:
// +---------+----------------+----------+
// |   ID    |       IP       |   NAME   |
// +---------+----------------+----------+
// |       1 | 180.17.95.2    | Jack     |
// |       2 | 180.17.95.3    | Wong     |
// |       3 | 180.17.95.4    | Lin      |
// |       4 | 180.17.95.5    | Trump    |
// +---------+----------------+----------+
func RowsFormater(rows *sql.Rows) {
	cols, err := rows.Columns()
	if err != nil {
		glog.Errorln(err)
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(cols)
	data := make([][]string, len(cols))

	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i, _ := range columns {
			columnPointers[i] = &columns[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(columnPointers...); err != nil {
			fmt.Println(err)
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		row := make([]string, 0)
		for i, _ := range cols {
			val := columnPointers[i].(*interface{})
			row = append(row, interface2String(*val))
		}
		data = append(data, row)
	}

	for _, v := range data {
		table.Append(v)
	}
	table.Render()
}

// Convert type interface{} into string just for friendly display.
func interface2String(input interface{}) string {
	switch input.(type) {
	case string:
		return input.(string)
	case int64:
		return strconv.FormatInt(input.(int64), 10)
	case []byte:
		return string(input.([]byte))
	default:
		return ""
	}
	return ""
}
