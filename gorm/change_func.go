package gorm

import (
	"fmt"
	"github.com/fangxing98/jx-gorm/gorm/clause"
	"github.com/fangxing98/jx-gorm/gorm/schema"
	"github.com/fangxing98/jx-gorm/gorm/utils"

	"strings"
)

// IsPgDriver 是否使用PG驱动
func (db *DB) IsPgDriver() bool {
	if db.DBType == DBTypePostgres || db.DBType == DBTypeKingBase {
		return true
	}

	return false
}

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

/*
关键字符号替换
s string: 原字符串
symbol string: 反引号需要替换的字符
all bool: 是否全部替换 false则不替换首尾
因为不同数据库类型关键字防止转义符号有所不同，但业务代码中统一使用的是MySQL方式的反引号，此处需要进行替换一下
eg

	mysql：关键使用反引号防止转义 `
	pg：关键字则使用双引号 "
*/
func (db *DB) keywordSymbolReplace(s string, all bool) string {

	if s == "" {
		return s
	}

	var symbol string

	if db.IsPgDriver() {
		symbol = "\""
	}

	if symbol == "" {
		return s
	}

	var firstStr, lastStr string

	if all {
		return strings.ReplaceAll(s, "`", symbol)
	}

	if len(s) >= 2 && s[0] == '`' && s[len(s)-1] == '`' {
		firstStr = "`"
		lastStr = "`"
		s = s[1 : len(s)-1]
	}

	s = strings.ReplaceAll(s, "`", symbol)

	if firstStr != "" {
		s = firstStr + s
	}

	if lastStr != "" {
		s = s + lastStr
	}

	return s
}

func (db *DB) Group(name string) (tx *DB) {

	name = db.keywordSymbolReplace(name, true)

	tx = db.getInstance()

	fields := strings.FieldsFunc(name, utils.IsValidDBNameChar)
	tx.Statement.AddClause(clause.GroupBy{
		Columns: []clause.Column{{Name: name, Raw: len(fields) != 1}},
	})
	return
}

func (db *DB) Having(query interface{}, args ...interface{}) (tx *DB) {

	switch v := query.(type) {
	case string:
		query = db.keywordSymbolReplace(v, true)
	}
	tx = db.getInstance()
	tx.Statement.AddClause(clause.GroupBy{
		Having: tx.Statement.BuildCondition(query, args...),
	})
	return
}

func (db *DB) Order(value interface{}) (tx *DB) {
	tx = db.getInstance()

	switch v := value.(type) {
	case clause.OrderBy:
		tx.Statement.AddClause(v)
	case clause.OrderByColumn:
		tx.Statement.AddClause(clause.OrderBy{
			Columns: []clause.OrderByColumn{v},
		})
	case string:
		v = db.keywordSymbolReplace(v, true)
		if v != "" {
			tx.Statement.AddClause(clause.OrderBy{
				Columns: []clause.OrderByColumn{{
					Column: clause.Column{Name: v, Raw: true},
				}},
			})
		}
	}
	return
}

func (db *DB) Select(query interface{}, args ...interface{}) (tx *DB) {
	tx = db.getInstance()

	switch v := query.(type) {
	case []string:
		tx.Statement.Selects = v

		for _, arg := range args {
			switch arg := arg.(type) {
			case string:
				tx.Statement.Selects = append(tx.Statement.Selects, arg)
			case []string:
				tx.Statement.Selects = append(tx.Statement.Selects, arg...)
			default:
				tx.AddError(fmt.Errorf("unsupported select args %v %v", query, args))
				return
			}
		}

		if clause, ok := tx.Statement.Clauses["SELECT"]; ok {
			clause.Expression = nil
			tx.Statement.Clauses["SELECT"] = clause
		}
	case string:

		v = db.keywordSymbolReplace(v, true)

		if strings.Count(v, "?") >= len(args) && len(args) > 0 {
			tx.Statement.AddClause(clause.Select{
				Distinct:   db.Statement.Distinct,
				Expression: clause.Expr{SQL: v, Vars: args},
			})
		} else if strings.Count(v, "@") > 0 && len(args) > 0 {
			tx.Statement.AddClause(clause.Select{
				Distinct:   db.Statement.Distinct,
				Expression: clause.NamedExpr{SQL: v, Vars: args},
			})
		} else {
			tx.Statement.Selects = []string{v}

			for _, arg := range args {
				switch arg := arg.(type) {
				case string:
					tx.Statement.Selects = append(tx.Statement.Selects, arg)
				case []string:
					tx.Statement.Selects = append(tx.Statement.Selects, arg...)
				default:
					tx.Statement.AddClause(clause.Select{
						Distinct:   db.Statement.Distinct,
						Expression: clause.Expr{SQL: v, Vars: args},
					})
					return
				}
			}

			if clause, ok := tx.Statement.Clauses["SELECT"]; ok {
				clause.Expression = nil
				tx.Statement.Clauses["SELECT"] = clause
			}
		}
	default:
		tx.AddError(fmt.Errorf("unsupported select args %v %v", query, args))
	}

	return
}

func (db *DB) Where(query interface{}, args ...interface{}) (tx *DB) {

	switch v := query.(type) {
	case string:
		query = db.keywordSymbolReplace(v, true)
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
	default:
		tx = db.getInstance()
		if conds := tx.Statement.BuildCondition(query, args...); len(conds) > 0 {
			tx.Statement.AddClause(clause.Where{Exprs: conds})
		}
	}

	return
}

func (db *DB) Raw(sql string, values ...interface{}) (tx *DB) {
	tx = db.getInstance()
	tx.Statement.SQL = strings.Builder{}

	sql = db.keywordSymbolReplace(sql, false)

	if strings.Contains(sql, "@") {
		clause.NamedExpr{SQL: sql, Vars: values}.Build(tx.Statement)
	} else {
		clause.Expr{SQL: sql, Vars: values}.Build(tx.Statement)
	}
	return
}

// Exec executes raw sql
func (db *DB) Exec(sql string, values ...interface{}) (tx *DB) {
	tx = db.getInstance()
	tx.Statement.SQL = strings.Builder{}

	sql = db.keywordSymbolReplace(sql, false)

	if strings.Contains(sql, "@") {
		clause.NamedExpr{SQL: sql, Vars: values}.Build(tx.Statement)
	} else {
		clause.Expr{SQL: sql, Vars: values}.Build(tx.Statement)
	}

	return tx.callbacks.Raw().Execute(tx)
}

/*
PgDBTypeMap pg数据库类型映射
*/
func PgDBTypeMap(tableName string, fieldsInfo *schema.Field) string {

	/*
		字段名
	*/
	fieldsName := fieldsInfo.DBName

	/*
		原字段类型  字母统一转小写进行判断
	*/
	oldType := strings.ToLower(string(fieldsInfo.DataType))

	/*
		描述：默认int id无自增功能
		解决方案：int类型主键设置为SERIAL自增
	*/
	if fieldsName == "id" && utils.Contains([]string{"int", "uint", "int64", "uint64"}, oldType) && fieldsInfo.PrimaryKey {
		return "SERIAL"
	}

	/*
		描述：pg模式 结构体字段为bool数据库字段为int会导致插入数据报错
		解决方案：判断结构体字段为bool时强制将数据库字段设置为bool
	*/
	if fieldsInfo.GORMDataType == "bool" {
		fmt.Printf("表：%s 字段：%s 数据库字段类型为int，结构体字段为bool PG驱动，已强制将数据库字段置为bool类型 \n", tableName, fieldsName)
		return "bool"
	}

	/*
		描述：pg模式不支持longtext字段类型
		解决方案：将longtext替换为text
	*/
	if oldType == "longtext" {
		fmt.Printf("表：%s 字段：%s PG驱动longtext类型，已自动转换为text类型 \n", tableName, fieldsName)
		return "text"
	}

	/*
		描述：pg模式不支持uint字段类型
		解决方案：将uint替换为int
	*/
	if oldType == "uint" {
		fmt.Printf("表：%s 字段：%s PG驱动uint类型，已自动转换为int类型 \n", tableName, fieldsName)
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
		fmt.Printf("表：%s 字段：%s PG驱动string类型，已自动转换为text类型 \n", tableName, fieldsName)
		return "text"
	}

	/*
		描述：pg模式不支持unsigned语法
		解决方案：去除 unsigned
	*/
	if oldType == "int unsigned" {
		fmt.Printf("表：%s 字段：%s PG驱动unsigned语法，已自动去除 \n", tableName, fieldsName)
		return "int"
	}

	/*
		描述：pg模式不支持tinyint类型
		解决方案：转int类型
	*/
	if strings.HasPrefix(oldType, "tinyint") {
		fmt.Printf("表：%s 字段：%s PG驱动tinyint类型，已自动转换为int类型 \n", tableName, fieldsName)
		return "int"
	}

	/*
		描述：pg模式 int类型不支持指定长度
		解决方案：自动去除长度
	*/
	if strings.HasPrefix(oldType, "int(") {
		fmt.Printf("表：%s 字段：%s PG驱动指定int长度，已自动去除 \n", tableName, fieldsName)
		return "int"
	}

	return oldType
}
