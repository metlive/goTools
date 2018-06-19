package backUp

import (
	"os"
	"fmt"
	"time"
	"strconv"
	"../common"
	"../db"
	"encoding/json"
	"github.com/zhoutk/jsonparser"
)

func ExportOne(fields common.DbConnFields, workDir string) {
	var fileName string
	if fields.FileAlias == "" {
		fileName = workDir + fields.DbName + "-" + time.Now().Format("2006-01-02") + ".sql"
	}else{
		fileName = workDir + fields.FileAlias + "-" + time.Now().Format("2006-01-02") + ".sql"
	}

	content := "/*   Mysql export" +
		"\n\n		Host: " + fields.DbHost +
		"\n\n		Port: " + strconv.Itoa(fields.DbPort) +
		"\n\n		DataBase: " + fields.DbName +
		"\n\n		Date: " + time.Now().Format("2006-01-02 15:04:05") +
		"\n\n		Author: zhoutk@189.cn" +
		"\n\n		Copyright: tlwl-2018" +
		"\n*/\n\n"
	writeToFile(fileName, content, false)
	writeToFile(fileName, "SET FOREIGN_KEY_CHECKS=0;\n\n", true)
	sqlStr := "select CONSTRAINT_NAME,TABLE_NAME,COLUMN_NAME,REFERENCED_TABLE_SCHEMA," +
		"REFERENCED_TABLE_NAME,REFERENCED_COLUMN_NAME from information_schema.`KEY_COLUMN_USAGE` " +
		"where REFERENCED_TABLE_SCHEMA = ? "
	var values []interface{}
	values = append(values, fields.DbName)
	rs, err := db.ExecuteWithDbConn(sqlStr, values, fields)
	if err != nil{
		fmt.Print(err)
		return
	}
	rows := rs["rows"].([]map[string]string)
	FKEYS := []byte(`{}`)
	d0 := []byte(``)
	for i := 0; i < len(rows); i++ {
		_, _, _, err := jsonparser.Get(FKEYS, rows[i]["TABLE_NAME"]+"."+rows[i]["CONSTRAINT_NAME"])
		if err != nil {
			value := []byte(`{"constraintName":` + rows[i]["CONSTRAINT_NAME"] + `,"sourceCols":["你好"],"schema":`+
				rows[i]["REFERENCED_TABLE_SCHEMA"]+`,"tableName":`+rows[i]["REFERENCED_TABLE_NAME"]+
				`,"targetCols":[]}`)
			d0,_ = jsonparser.Set(FKEYS, value,rows[i]["TABLE_NAME"]+"."+rows[i]["CONSTRAINT_NAME"])
		}
		d, _, _, err := jsonparser.Get(d0, rows[i]["TABLE_NAME"]+"."+rows[i]["CONSTRAINT_NAME"], "sourceCols")
		d1 := toStringArray(d)
		d1 = append(d1, rows[i]["COLUMN_NAME"])

		FKEYS,_ = jsonparser.Set(d0, stringArrToByteArr(d1), rows[i]["TABLE_NAME"]+"."+rows[i]["CONSTRAINT_NAME"], "sourceCols")
		//FKEYS[rows[i]["TABLE_NAME"]+"."+rows[i]["CONSTRAINT_NAME"]].(map[string]interface{})["sourceCols"] =
		//	append(FKEYS[rows[i]["TABLE_NAME"]+"."+rows[i]["CONSTRAINT_NAME"]].(map[string]interface{})["sourceCols"].([]interface{}), rows[i]["COLUMN_NAME"])
		//FKEYS[rows[i]["TABLE_NAME"]+"."+rows[i]["CONSTRAINT_NAME"]].(map[string]interface{})["targetCols"] =
		//	append(FKEYS[rows[i]["TABLE_NAME"]+"."+rows[i]["CONSTRAINT_NAME"]].(map[string]interface{})["targetCols"].([]interface{}), rows[i]["COLUMN_NAME"])
	}
	data, _ := json.Marshal(FKEYS)
	fmt.Print(string(data))
}

func stringArrToByteArr(str []string) (x []byte) {
	x = append(x, '[')
	for i:=0; i<len(str); i++{
		b := []rune(str[i])
		x = append(x, '"')
			tmp := []byte(string(b))
			for j:=0; j < len(tmp); j++ {
				x = append(x, tmp[j])
			}
		x = append(x, '"')
		if i < len(str) -1 {
			x = append(x, ',')
		}
	}
	x = append(x, ']')
	return
}

func toStringArray(data []byte) (result []string) {
	jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		result = append(result, string(value))
	})
	return
}

func writeToFile(name string, content string, append bool)  {
	var fileObj *os.File
	var err error
	if append{
		fileObj, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	}else{
		fileObj, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	}
	if err != nil {
		fmt.Println("Failed to open the file", err.Error())
		os.Exit(2)
	}
	defer fileObj.Close()
	if _, err := fileObj.WriteString(content); err == nil {
		fmt.Println("Successful writing to the file.")
	}
}