package soft_delete_test

import (
	"errors"
	"github.com/fangxing98/jx-gorm/driver/sqlite"
	"github.com/fangxing98/jx-gorm/gorm"
	"github.com/fangxing98/jx-gorm/plugin/soft_delete"

	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"
)

type User struct {
	ID        uint
	Name      string
	Age       uint
	DeletedAt soft_delete.DeletedAt
}

func TestSoftDelete(t *testing.T) {
	DB, err := gorm.Open(sqlite.Open(filepath.Join(os.TempDir(), "gorm.db")), gorm.DBTypeSqlite, &gorm.Config{})
	DB = DB.Debug()
	if err != nil {
		t.Errorf("failed to connect database")
	}

	user := User{Name: "jinzhu", Age: 20}
	DB.Migrator().DropTable(&User{})
	DB.AutoMigrate(&User{})
	DB.Save(&user)

	var count int64
	var age uint

	if DB.Model(&User{}).Where("name = ?", user.Name).Count(&count).Error != nil || count != 1 {
		t.Errorf("Count soft deleted record, expects: %v, got: %v", 1, count)
	}

	if DB.Model(&User{}).Select("age").Where("name = ?", user.Name).Scan(&age).Error != nil || age != user.Age {
		t.Errorf("Age soft deleted record, expects: %v, got: %v", 0, age)
	}

	if err := DB.Delete(&user).Error; err != nil {
		t.Fatalf("No error should happen when soft delete user, but got %v", err)
	}

	if user.DeletedAt == 0 {
		t.Errorf("user's deleted at should not be zero, DeletedAt: %v", user.DeletedAt)
	}

	sql := DB.Session(&gorm.Session{DryRun: true}).Delete(&user).Statement.SQL.String()
	if !regexp.MustCompile(`UPDATE .users. SET .deleted_at.=.* WHERE .users.\..id. = .* AND .users.\..deleted_at. = ?`).MatchString(sql) {
		t.Fatalf("invalid sql generated, got %v", sql)
	}

	if DB.First(&User{}, "name = ?", user.Name).Error == nil {
		t.Errorf("Can't find a soft deleted record")
	}

	count = 0
	if DB.Model(&User{}).Where("name = ?", user.Name).Count(&count).Error != nil || count != 0 {
		t.Errorf("Count soft deleted record, expects: %v, got: %v", 0, count)
	}

	age = 0
	if err := DB.Model(&User{}).Select("age").Where("name = ?", user.Name).Scan(&age).Error; err != nil || age != 0 {
		t.Fatalf("Age soft deleted record, expects: %v, got: %v, err %v", 0, age, err)
	}

	if err := DB.Unscoped().First(&User{}, "name = ?", user.Name).Error; err != nil {
		t.Errorf("Should find soft deleted record with Unscoped, but got err %s", err)
	}

	count = 0
	if DB.Unscoped().Model(&User{}).Where("name = ?", user.Name).Count(&count).Error != nil || count != 1 {
		t.Errorf("Count soft deleted record, expects: %v, count: %v", 1, count)
	}

	age = 0
	if DB.Unscoped().Model(&User{}).Select("age").Where("name = ?", user.Name).Scan(&age).Error != nil || age != user.Age {
		t.Errorf("Age soft deleted record, expects: %v, got: %v", 0, age)
	}

	DB.Unscoped().Delete(&user)
	if err := DB.Unscoped().First(&User{}, "name = ?", user.Name).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("Can't find permanently deleted record")
	}
}

type MilliUser struct {
	ID        uint
	Name      string
	Age       uint
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
}

func TestSoftDeleteMilliMode(t *testing.T) {
	DB, err := gorm.Open(sqlite.Open(filepath.Join(os.TempDir(), "gorm.db")), gorm.DBTypeSqlite, &gorm.Config{})
	DB = DB.Debug()
	if err != nil {
		t.Errorf("failed to connect database")
	}

	user := MilliUser{Name: "jinzhu", Age: 20}
	DB.Migrator().DropTable(&MilliUser{})
	DB.AutoMigrate(&MilliUser{})
	DB.Save(&user)

	var count int64
	var age uint

	if DB.Model(&MilliUser{}).Where("name = ?", user.Name).Count(&count).Error != nil || count != 1 {
		t.Errorf("Count soft deleted record, expects: %v, got: %v", 1, count)
	}

	if DB.Model(&MilliUser{}).Select("age").Where("name = ?", user.Name).Scan(&age).Error != nil || age != user.Age {
		t.Errorf("Age soft deleted record, expects: %v, got: %v", 0, age)
	}

	if err := DB.Delete(&user).Error; err != nil {
		t.Fatalf("No error should happen when soft delete user, but got %v", err)
	}

	if user.DeletedAt == 0 {
		t.Errorf("user's deleted at should not be zero, DeletedAt: %v", user.DeletedAt)
	}

	sql := DB.Session(&gorm.Session{DryRun: true}).Delete(&user).Statement.SQL.String()
	if !regexp.MustCompile(`UPDATE .milli_users. SET .deleted_at.=.* WHERE .milli_users.\..id. = .* AND .milli_users.\..deleted_at. = ?`).MatchString(sql) {
		t.Fatalf("invalid sql generated, got %v", sql)
	}

	if DB.First(&MilliUser{}, "name = ?", user.Name).Error == nil {
		t.Errorf("Can't find a soft deleted record")
	}

	count = 0
	if DB.Model(&MilliUser{}).Where("name = ?", user.Name).Count(&count).Error != nil || count != 0 {
		t.Errorf("Count soft deleted record, expects: %v, got: %v", 0, count)
	}

	age = 0
	if err := DB.Model(&MilliUser{}).Select("age").Where("name = ?", user.Name).Scan(&age).Error; err != nil || age != 0 {
		t.Fatalf("Age soft deleted record, expects: %v, got: %v, err %v", 0, age, err)
	}

	if err := DB.Unscoped().First(&MilliUser{}, "name = ?", user.Name).Error; err != nil {
		t.Errorf("Should find soft deleted record with Unscoped, but got err %s", err)
	}

	count = 0
	if DB.Unscoped().Model(&MilliUser{}).Where("name = ?", user.Name).Count(&count).Error != nil || count != 1 {
		t.Errorf("Count soft deleted record, expects: %v, count: %v", 1, count)
	}

	age = 0
	if DB.Unscoped().Model(&MilliUser{}).Select("age").Where("name = ?", user.Name).Scan(&age).Error != nil || age != user.Age {
		t.Errorf("Age soft deleted record, expects: %v, got: %v", 0, age)
	}

	DB.Unscoped().Delete(&user)
	if err := DB.Unscoped().First(&MilliUser{}, "name = ?", user.Name).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("Can't find permanently deleted record")
	}
}

type FlagUser struct {
	ID    uint
	Name  string
	Age   uint
	IsDel soft_delete.DeletedAt `gorm:"softDelete:flag"`
}

func TestSoftDeleteFlagMode(t *testing.T) {
	DB, err := gorm.Open(sqlite.Open(filepath.Join(os.TempDir(), "gorm.db")), gorm.DBTypeSqlite, &gorm.Config{})
	DB = DB.Debug()
	if err != nil {
		t.Errorf("failed to connect database")
	}

	user := FlagUser{Name: "jinzhu", Age: 20}
	DB.Migrator().DropTable(&FlagUser{})
	DB.AutoMigrate(&FlagUser{})
	DB.Save(&user)

	var count int64
	var age uint

	if DB.Model(&FlagUser{}).Where("name = ?", user.Name).Count(&count).Error != nil || count != 1 {
		t.Errorf("Count soft deleted record, expects: %v, got: %v", 1, count)
	}

	if DB.Model(&FlagUser{}).Select("age").Where("name = ?", user.Name).Scan(&age).Error != nil || age != user.Age {
		t.Errorf("Age soft deleted record, expects: %v, got: %v", 0, age)
	}

	if err := DB.Delete(&user).Error; err != nil {
		t.Fatalf("No error should happen when soft delete user, but got %v", err)
	}

	if user.IsDel == 0 {
		t.Errorf("user's deleted at should not be zero, IsDel: %v", user.IsDel)
	}

	sql := DB.Session(&gorm.Session{DryRun: true}).Delete(&user).Statement.SQL.String()
	if !regexp.MustCompile(`UPDATE .flag_users. SET .is_del.=.* WHERE .flag_users.\..id. = .* AND .flag_users.\..is_del. = ?`).MatchString(sql) {
		t.Fatalf("invalid sql generated, got %v", sql)
	}

	if DB.First(&FlagUser{}, "name = ?", user.Name).Error == nil {
		t.Errorf("Can't find a soft deleted record")
	}

	count = 0
	if DB.Model(&FlagUser{}).Where("name = ?", user.Name).Count(&count).Error != nil || count != 0 {
		t.Errorf("Count soft deleted record, expects: %v, got: %v", 0, count)
	}

	age = 0
	if err := DB.Model(&FlagUser{}).Select("age").Where("name = ?", user.Name).Scan(&age).Error; err != nil || age != 0 {
		t.Fatalf("Age soft deleted record, expects: %v, got: %v, err %v", 0, age, err)
	}

	if err := DB.Unscoped().First(&FlagUser{}, "name = ?", user.Name).Error; err != nil {
		t.Errorf("Should find soft deleted record with Unscoped, but got err %s", err)
	}

	count = 0
	if DB.Unscoped().Model(&FlagUser{}).Where("name = ?", user.Name).Count(&count).Error != nil || count != 1 {
		t.Errorf("Count soft deleted record, expects: %v, count: %v", 1, count)
	}

	age = 0
	if DB.Unscoped().Model(&FlagUser{}).Select("age").Where("name = ?", user.Name).Scan(&age).Error != nil || age != user.Age {
		t.Errorf("Age soft deleted record, expects: %v, got: %v", 0, age)
	}

	DB.Unscoped().Delete(&user)
	if err := DB.Unscoped().First(&FlagUser{}, "name = ?", user.Name).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("Can't find permanently deleted record")
	}
}

type MixedUser struct {
	ID        uint
	Name      string
	Age       uint
	DeletedAt time.Time
	IsDel     soft_delete.DeletedAt `gorm:"softDelete:flag,DeletedAtField:DeletedAt"`
}

type MixedUser2 struct {
	ID        uint
	Name      string
	Age       uint
	DeletedAt time.Time
	IsDel     soft_delete.DeletedAt `gorm:"softDelete:,DeletedAtField:DeletedAt"`
}

func TestMixedDeleteFlagMode(t *testing.T) {
	DB, err := gorm.Open(sqlite.Open(filepath.Join(os.TempDir(), "gorm.db")), gorm.DBTypeSqlite, &gorm.Config{})
	DB = DB.Debug()
	if err != nil {
		t.Errorf("failed to connect database")
	}

	user := MixedUser{Name: "jinzhu", Age: 20}
	DB.Migrator().DropTable(&MixedUser{})
	DB.AutoMigrate(&MixedUser{})
	DB.Save(&user)

	var count int64
	var age uint

	if DB.Model(&MixedUser{}).Where("name = ?", user.Name).Count(&count).Error != nil || count != 1 {
		t.Errorf("Count soft deleted record, expects: %v, got: %v", 1, count)
	}

	if DB.Model(&MixedUser{}).Select("age").Where("name = ?", user.Name).Scan(&age).Error != nil || age != user.Age {
		t.Errorf("Age soft deleted record, expects: %v, got: %v", 0, age)
	}

	if err := DB.Delete(&user).Error; err != nil {
		t.Fatalf("No error should happen when soft delete user, but got %v", err)
	}

	if user.DeletedAt.IsZero() || user.IsDel != 1 {
		t.Errorf("user's deleted at should not be zero, DeletedAt: %v, IsDel: %v", user.DeletedAt, user.IsDel)
	}

	sql := DB.Session(&gorm.Session{DryRun: true}).Delete(&user).Statement.SQL.String()
	if !regexp.MustCompile(`UPDATE .mixed_users. SET .is_del.=.*,.deleted_at.=.* WHERE .mixed_users.\..id. = .* AND .mixed_users.\..is_del. = ?`).MatchString(sql) {
		t.Fatalf("invalid sql generated, got %v", sql)
	}

	if DB.First(&MixedUser{}, "name = ?", user.Name).Error == nil {
		t.Errorf("Can't find a soft deleted record")
	}

	count = 0
	if DB.Model(&MixedUser{}).Where("name = ?", user.Name).Count(&count).Error != nil || count != 0 {
		t.Errorf("Count soft deleted record, expects: %v, got: %v", 0, count)
	}

	age = 0
	if err := DB.Model(&MixedUser{}).Select("age").Where("name = ?", user.Name).Scan(&age).Error; err != nil || age != 0 {
		t.Fatalf("Age soft deleted record, expects: %v, got: %v, err %v", 0, age, err)
	}

	if err := DB.Unscoped().First(&MixedUser{}, "name = ?", user.Name).Error; err != nil {
		t.Errorf("Should find soft deleted record with Unscoped, but got err %s", err)
	}

	count = 0
	if DB.Unscoped().Model(&MixedUser{}).Where("name = ?", user.Name).Count(&count).Error != nil || count != 1 {
		t.Errorf("Count soft deleted record, expects: %v, count: %v", 1, count)
	}

	age = 0
	if DB.Unscoped().Model(&MixedUser{}).Select("age").Where("name = ?", user.Name).Scan(&age).Error != nil || age != user.Age {
		t.Errorf("Age soft deleted record, expects: %v, got: %v", 0, age)
	}

	DB.Unscoped().Delete(&user)

	if err := DB.Unscoped().First(&MixedUser{}, "name = ?", user.Name).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("Can't find permanently deleted record")
	}

	user2 := MixedUser2{Name: "jinzhu", Age: 20}
	DB.Table("mixed_users").Save(&user2)

	if err := DB.Table("mixed_users").Delete(&user2).Error; err != nil {
		t.Fatalf("No error should happen when soft delete user, but got %v", err)
	}

	if user2.DeletedAt.IsZero() || int64(user2.IsDel) <= time.Now().Unix()-10 {
		t.Errorf("user's deleted at should not be zero, DeletedAt: %v, IsDel: %v", user2.DeletedAt, user2.IsDel)
	}
}

type NullableDeletedAtUser struct {
	ID        int64
	Name      string
	Age       uint
	DeletedAt soft_delete.DeletedAt `gorm:"default:null"`
}

func TestNullableDeletedAtUser(t *testing.T) {
	DB, err := gorm.Open(sqlite.Open(filepath.Join(os.TempDir(), "gorm.db")), gorm.DBTypeSqlite, &gorm.Config{})
	DB = DB.Debug()
	if err != nil {
		t.Errorf("failed to connect database")
	}

	user := NullableDeletedAtUser{Name: "shyamin", Age: 25}
	DB.Migrator().DropTable(&NullableDeletedAtUser{})
	DB.AutoMigrate(&NullableDeletedAtUser{})
	DB.Save(&user)

	var count int64
	var age uint

	if DB.Model(&NullableDeletedAtUser{}).Where("name = ?", user.Name).Count(&count).Error != nil || count != 1 {
		t.Errorf("Count soft deleted record, expects: %v, got: %v", 1, count)
	}

	if DB.Model(&NullableDeletedAtUser{}).Select("age").Where("name = ?", user.Name).Scan(&age).Error != nil || age != user.Age {
		t.Errorf("Age soft deleted record, expects: %v, got: %v", 0, age)
	}

	if err := DB.Delete(&user).Error; err != nil {
		t.Fatalf("No error should happen when soft delete user, but got %v", err)
	}

}
