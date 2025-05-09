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
