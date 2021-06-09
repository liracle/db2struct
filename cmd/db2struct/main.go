package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Shelnutt2/db2struct"
	"github.com/droundy/goopt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/howeyc/gopass"
)

var mariadbHost = os.Getenv("MYSQL_HOST")
var mariadbHostPassed = goopt.String([]string{"-H", "--host"}, "", "Host to check mariadb status of")
var mariadbPort = goopt.Int([]string{"--mysql_port"}, 3306, "Specify a port to connect to")
var mariadbTable = goopt.String([]string{"-t", "--table"}, "", "Table list to build struct from,separate by comma")
var mariadbDatabase = goopt.String([]string{"-d", "--database"}, "nil", "Database to for connection")
var mariadbPassword *string
var mariadbUser = goopt.String([]string{"-u", "--user"}, "user", "user to connect to database")
var verbose = goopt.Flag([]string{"-v", "--verbose"}, []string{}, "Enable verbose output", "")
var noSqlType = goopt.Flag([]string{"--nosqltype"}, []string{}, "生成的sql不含有sqltype", "")

//var packageName = goopt.String([]string{"--package"}, "", "name to set for package")
//var structName = goopt.String([]string{"--struct"}, "", "name to set for struct")

var jsonAnnotation = goopt.Flag([]string{"--json"}, []string{"--no-json"}, "Add json annotations (default)", "Disable json annotations")
var gormAnnotation = goopt.Flag([]string{"--gorm"}, []string{}, "Add gorm annotations (tags)", "")
var gureguTypes = goopt.Flag([]string{"--guregu"}, []string{}, "Add guregu null types", "")
var targetFile = goopt.String([]string{"--target"}, "", "Save file path")

func init() {
	goopt.OptArg([]string{"-p", "--password"}, "", "Mysql password", getMariadbPassword)
	//goopt.ReqArg([]string{"-u", "--user"}, "user", "user to connect to database", setUser)

	// Setup goopts
	goopt.Description = func() string {
		return "Mariadb http Check"
	}
	goopt.Version = "0.0.3"
	goopt.Summary = "db2struct [-H] [-p] [-v] --database databaseName --table tableName"

	//Parse options
	goopt.Parse(nil)

}

func main() {

	// Username is required
	if mariadbUser == nil || *mariadbUser == "user" {
		fmt.Println("Username is required! Add it with --user=name")
		return
	}

	// If a mariadb host is passed use it
	if mariadbHostPassed != nil && *mariadbHostPassed != "" {
		mariadbHost = *mariadbHostPassed
	}

	if mariadbPassword != nil && *mariadbPassword == "" {
		fmt.Print("Password: ")
		pass, err := gopass.GetPasswd()
		stringPass := string(pass)
		mariadbPassword = &stringPass
		if err != nil {
			fmt.Println("Error reading password: " + err.Error())
			return
		}
	} else if mariadbPassword == nil {
		p := ""
		mariadbPassword = &p
	}

	if *verbose {
		fmt.Println("Connecting to mysql server " + mariadbHost + ":" + strconv.Itoa(*mariadbPort))
	}

	if mariadbDatabase == nil || *mariadbDatabase == "" {
		fmt.Println("Database can not be null")
		return
	}

	if mariadbTable == nil || *mariadbTable == "" {
		fmt.Println("Table can not be null")
		return
	}
	tables := strings.Split(*mariadbTable, ",")
	lines := make([]string, 0, len(tables))
	for _, table := range tables {
		columnDataTypes, columnsSorted, err := db2struct.GetColumnsFromMysqlTable(*mariadbUser, *mariadbPassword, mariadbHost, *mariadbPort, *mariadbDatabase, table)

		if err != nil {
			fmt.Println("Error in selecting column data information from mysql information schema")
			return
		}
		structName := Trans2StructName(table)

		// Generate struct string based on columnDataTypes
		struc, err := db2struct.Generate(*columnDataTypes, columnsSorted, table, structName, *jsonAnnotation, *gormAnnotation, *gureguTypes, *noSqlType)

		if err != nil {
			fmt.Println("Error in creating struct from json: " + err.Error())
			return
		}
		lines = append(lines, string(struc))
	}
	if targetFile != nil && *targetFile != "" {
		file, err := os.OpenFile(*targetFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Open File fail: " + err.Error())
			return
		}
		length, err := file.WriteString(strings.Join(lines, ""))
		if err != nil {
			fmt.Println("Save File fail: " + err.Error())
			return
		}
		fmt.Printf("wrote %d bytes\n", length)
	} else {
		fmt.Println(strings.Join(lines, ""))
	}
}

func Trans2StructName(tableName string) string {
	if tableName == "" {
		return ""
	}
	if strings.HasPrefix(tableName, "t_") {
		tableName = tableName[2:]
	}
	tableName = strings.Replace(tableName, "_", " ", -1)
	tableName = strings.Title(tableName)
	return strings.Replace(tableName, " ", "", -1)
}

func getMariadbPassword(password string) error {
	mariadbPassword = new(string)
	*mariadbPassword = password
	return nil
}
