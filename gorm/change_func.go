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
*/
func PgDBTypeMap(tableName string, fieldsInfo *schema.Field) string {

	fieldsName := fieldsInfo.DBName        // 字段名
	oldType := string(fieldsInfo.DataType) // 原字段类型

	/*
		描述：默认int id无自增功能
		解决方案：int类型主键设置为SERIAL自增
	*/
	if fieldsName == "id" && utils.Contains([]string{"int", "uint"}, oldType) {
		return "SERIAL"
	}

	/*
		描述：pg模式 结构体字段为bool数据库字段为int会导致插入数据报错
		解决方案：判断结构体字段为bool时强制将数据库字段设置为bool
	*/
	if fieldsInfo.GORMDataType == "bool" {
		fmt.Printf("表：%s 字段：%s 数据库字段类型为int，结构体字段为bool kingbase不支持，已强制将数据库字段置为bool类型 \n", tableName, fieldsName)
		return "bool"
	}

	/*
		描述：pg模式不支持longtext字段类型
		解决方案：将longtext替换为text
	*/
	if oldType == "longtext" {
		fmt.Printf("表：%s 字段：%s kingbase不支持longtext类型，已自动转换为text类型 \n", tableName, fieldsName)
		return "text"
	}

	/*
		描述：pg模式不支持uint字段类型
		解决方案：将uint替换为int
	*/
	if oldType == "uint" {
		fmt.Printf("表：%s 字段：%s kingbase不支持uint类型，已自动转换为int类型 \n", tableName, fieldsName)
		return "int"
	}

	/*
		描述：int32 int64 默认转换类型int4太小不符合需求，例如13位时间戳  size为位数 int64 size为64 int32 size为32
		解决方案：使用bigint类型替换
	*/
	if oldType == "int" && fieldsInfo.Size >= 32 {
		return "bigint"
	}

	/*
		描述：pg模式不支持string字段类型
		解决方案：将string替换为text
	*/
	if oldType == "string" {
		fmt.Printf("表：%s 字段：%s kingbase不支持string类型，已自动转换为text类型 \n", tableName, fieldsName)
		return "text"
	}

	/*
		描述：pg模式不支持unsigned语法
		解决方案：去除 unsigned
	*/
	if oldType == "int unsigned" {
		fmt.Printf("表：%s 字段：%s kingbase不支持unsigned语法，已自动去除 \n", tableName, fieldsName)
		return "int"
	}

	/*
		描述：pg模式不支持tinyint类型
		解决方案：转int类型
	*/
	if strings.HasPrefix(oldType, "tinyint") {
		fmt.Printf("表：%s 字段：%s kingbase不支持tinyint类型，已自动转换为int类型 \n", tableName, fieldsName)
		return "int"
	}

	/*
		描述：pg模式 int类型不支持指定长度
		解决方案：自动去除长度
	*/
	if strings.HasPrefix(oldType, "int(") {
		fmt.Printf("表：%s 字段：%s kingbase不支持指定int长度，已自动去除 \n", tableName, fieldsName)
		return "int"
	}

	return oldType
}
