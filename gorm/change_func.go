package gorm

import (
	"fmt"
	"github.com/fangxing98/jx-gorm/gorm/clause"
	"github.com/fangxing98/jx-gorm/gorm/schema"
	"github.com/fangxing98/jx-gorm/gorm/utils"

	"strings"
)

func (db *DB) replacePlaceholders(query string, args []interface{}) string {
	// 分割 SQL 查询语句
	placeholders := strings.Split(query, "?")
	var result strings.Builder

	// 遍历并替换
	for i, part := range placeholders {
		result.WriteString(part)
		if i < len(args) {
			// 根据类型替换 ?
			switch args[i].(type) {
			case string:
				result.WriteString("%s")
			case int:
				result.WriteString("%d")
			default:
				result.WriteString("%v")
			}
		}
	}

	return fmt.Sprintf(result.String(), args...)
}

func (db *DB) Where(query interface{}, args ...interface{}) (tx *DB) {

	if db.DBType == DBTypeKingBase {

		queryStr := query.(string)

		// 处理kingbase < > != 比对时可能失效问题
		if utils.Contains([]string{"<", ">", "!="}, queryStr) {
			newQuery := db.replacePlaceholders(queryStr, args)

			tx = db.getInstance()
			if conds := tx.Statement.BuildCondition(newQuery); len(conds) > 0 {
				tx.Statement.AddClause(clause.Where{Exprs: conds})
			}
			return
		}
	}

	tx = db.getInstance()
	if conds := tx.Statement.BuildCondition(query, args...); len(conds) > 0 {
		tx.Statement.AddClause(clause.Where{Exprs: conds})
	}
	return
}

/*
PgDBTypeMap pg(kingbase)数据库类型映射

场景1

	描述：pg模式不支持longtext字段类型
	解决方案：将longtext替换为text

场景2

	描述：pg模式不支持tinyint字段类型
	解决方案：将tinyint替换为int

场景3

	描述：pg模式 int类型不支持指定长度
	解决方案：自动去除长度

场景4

	描述：pg模式 结构体字段为bool数据库字段为int会导致插入数据报错
	解决方案：判断结构体字段为bool时强制将数据库字段设置为bool
*/
func PgDBTypeMap(tableName string, fieldsInfo *schema.Field) string {

	fieldsName := fieldsInfo.DBName        // 字段名
	oldType := string(fieldsInfo.DataType) // 原字段类型

	if fieldsInfo.GORMDataType == "bool" {
		fmt.Printf("表：%s 字段：%s 数据库字段类型为int，结构体字段为bool kingbase不支持，已强制将数据库字段置为bool类型 \n", tableName, fieldsName)
		return "bool"
	}

	if oldType == "longtext" {
		fmt.Printf("表：%s 字段：%s kingbase不支持longtext类型，已自动转换为text类型 \n", tableName, fieldsName)
		return "text"
	}

	// tinyint、tinyint(1) 处理
	if strings.HasPrefix(oldType, "tinyint") {
		fmt.Printf("表：%s 字段：%s kingbase不支持tinyint类型，已自动转换为int类型 \n", tableName, fieldsName)
		return "int"
	}

	// int(1) 处理
	if strings.HasPrefix(oldType, "int(") {
		fmt.Printf("表：%s 字段：%s kingbase不支持指定int长度，已自动去除 \n", tableName, fieldsName)
		return "int"
	}

	return oldType
}
