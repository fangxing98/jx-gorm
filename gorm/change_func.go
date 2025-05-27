package gorm

import (
	"fmt"
	"github.com/fangxing98/jx-gorm/gorm/clause"
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

1.不支持类型：longtext tinyint
2.int不支持指定长度
*/
func PgDBTypeMap(oldType, fieldsName, tableName string) string {

	if oldType == "longtext" {
		fmt.Printf("表：%s 字段：%s pg 模式 kingbase 不支持longtext类型，已自动转换为text类型 \n", tableName, fieldsName)
		return "text"
	}

	if oldType == "tinyint" {
		fmt.Printf("表：%s 字段：%s pg 模式 kingbase 不支持tinyint类型，已自动转换为int类型 \n", tableName, fieldsName)
		return "int"
	}

	if strings.HasPrefix(oldType, "int(") {
		fmt.Printf("表：%s 字段：%s pg 模式 kingbase 不支持指定int长度，已自动去除 \n", tableName, fieldsName)
		return "int"
	}

	return oldType
}
