package sqlserver

import (
	"github.com/fangxing98/jx-gorm/gorm"
	"github.com/fangxing98/jx-gorm/gorm/callbacks"
)

var updateFunc = callbacks.Update(&callbacks.Config{
	UpdateClauses: []string{"UPDATE", "SET", "RETURNING", "FROM", "WHERE"},
})

func Update(db *gorm.DB) {
	if db.Statement.Schema != nil && db.Statement.Schema.PrioritizedPrimaryField != nil && db.Statement.Schema.PrioritizedPrimaryField.AutoIncrement {
		db.Statement.Omits = append(db.Statement.Omits, db.Statement.Schema.PrioritizedPrimaryField.DBName)
	}

	updateFunc(db)
}
