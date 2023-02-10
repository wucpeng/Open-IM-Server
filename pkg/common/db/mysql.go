package db

import (
	"Open_IM/pkg/common/config"
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type mysqlDB struct {
	sync.RWMutex
	db *gorm.DB
}

type Writer struct{}

func (w Writer) Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func initMysqlDB() {
	fmt.Println("init mysqlDB start")
	//When there is no open IM database, connect to the mysql built-in database to create openIM database
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		config.Config.Mysql.DBUserName, config.Config.Mysql.DBPassword, config.Config.Mysql.DBAddress[0], "mysql")
	var db *gorm.DB
	var err1 error
	db, err := gorm.Open(mysql.Open(dsn), nil)
	if err != nil {
		fmt.Println("Open failed ", err.Error(), dsn)
	}
	if err != nil {
		time.Sleep(time.Duration(30) * time.Second)
		db, err1 = gorm.Open(mysql.Open(dsn), nil)
		if err1 != nil {
			fmt.Println("Open failed ", err1.Error(), dsn)
			panic(err1.Error())
		}
	}

	//Check the database and table during initialization
	sql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s default charset utf8 COLLATE utf8_general_ci;", config.Config.Mysql.DBDatabaseName)
	fmt.Println("exec sql: ", sql, " begin")
	err = db.Exec(sql).Error
	if err != nil {
		fmt.Println("Exec failed ", err.Error(), sql)
		panic(err.Error())
	}
	fmt.Println("exec sql: ", sql, " end")
	dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		config.Config.Mysql.DBUserName, config.Config.Mysql.DBPassword, config.Config.Mysql.DBAddress[0], config.Config.Mysql.DBDatabaseName)

	newLogger := logger.New(
		Writer{},
		logger.Config{
			SlowThreshold:             time.Duration(config.Config.Mysql.SlowThreshold) * time.Millisecond, // Slow SQL threshold
			LogLevel:                  logger.LogLevel(config.Config.Mysql.LogLevel),                       // Log level
			IgnoreRecordNotFoundError: true,                                                                // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,                                                                // Disable color
		},
	)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		fmt.Println("Open failed ", err.Error(), dsn)
		panic(err.Error())
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(err.Error())
	}

	sqlDB.SetConnMaxLifetime(time.Second * time.Duration(config.Config.Mysql.DBMaxLifeTime))
	sqlDB.SetMaxOpenConns(config.Config.Mysql.DBMaxOpenConns)
	sqlDB.SetMaxIdleConns(config.Config.Mysql.DBMaxIdleConns)

	fmt.Println("open mysql ok ", dsn)
	db.AutoMigrate(
		&Register{},
		&Friend{},
		&FriendRequest{},
		&Group{},
		&GroupMember{},
		&GroupRequest{},
		&User{},
		&Black{}, &ChatLog{}, &Register{}, &Conversation{}, &AppVersion{}, &Department{}, &BlackList{}, &IpLimit{}, &UserIpLimit{}, &Invitation{}, &RegisterAddFriend{},
		&ClientInitConfig{}, &UserIpRecord{})
	db.Set("gorm:table_options", "CHARSET=utf8")
	db.Set("gorm:table_options", "collation=utf8_unicode_ci")

	if !db.Migrator().HasTable(&Friend{}) {
		fmt.Println("CreateTable Friend")
		db.Migrator().CreateTable(&Friend{})
	}

	if !db.Migrator().HasTable(&FriendRequest{}) {
		fmt.Println("CreateTable FriendRequest")
		db.Migrator().CreateTable(&FriendRequest{})
	}

	if !db.Migrator().HasTable(&Group{}) {
		fmt.Println("CreateTable Group")
		db.Migrator().CreateTable(&Group{})
	}

	if !db.Migrator().HasTable(&GroupMember{}) {
		fmt.Println("CreateTable GroupMember")
		db.Migrator().CreateTable(&GroupMember{})
	}
	if !db.Migrator().HasTable(&GroupRequest{}) {
		fmt.Println("CreateTable GroupRequest")
		db.Migrator().CreateTable(&GroupRequest{})
	}
	if !db.Migrator().HasTable(&User{}) {
		fmt.Println("CreateTable User")
		db.Migrator().CreateTable(&User{})
	}
	if !db.Migrator().HasTable(&Black{}) {
		fmt.Println("CreateTable Black")
		db.Migrator().CreateTable(&Black{})
	}
	if !db.Migrator().HasTable(&ChatLog{}) {
		fmt.Println("CreateTable ChatLog")
		db.Migrator().CreateTable(&ChatLog{})
	}
	if !db.Migrator().HasTable(&Register{}) {
		fmt.Println("CreateTable Register")
		db.Migrator().CreateTable(&Register{})
	}
	if !db.Migrator().HasTable(&Conversation{}) {
		fmt.Println("CreateTable Conversation")
		db.Migrator().CreateTable(&Conversation{})
	}

	if !db.Migrator().HasTable(&Department{}) {
		fmt.Println("CreateTable Department")
		db.Migrator().CreateTable(&Department{})
	}
	if !db.Migrator().HasTable(&OrganizationUser{}) {
		fmt.Println("CreateTable OrganizationUser")
		db.Migrator().CreateTable(&OrganizationUser{})
	}
	if !db.Migrator().HasTable(&DepartmentMember{}) {
		fmt.Println("CreateTable DepartmentMember")
		db.Migrator().CreateTable(&DepartmentMember{})
	}
	if !db.Migrator().HasTable(&AppVersion{}) {
		fmt.Println("CreateTable DepartmentMember")
		db.Migrator().CreateTable(&AppVersion{})
	}
	if !db.Migrator().HasTable(&BlackList{}) {
		fmt.Println("CreateTable BlackList")
		db.Migrator().CreateTable(&BlackList{})
	}
	if !db.Migrator().HasTable(&IpLimit{}) {
		fmt.Println("CreateTable IpLimit")
		db.Migrator().CreateTable(&IpLimit{})
	}
	if !db.Migrator().HasTable(&UserIpLimit{}) {
		fmt.Println("CreateTable UserIpLimit")
		db.Migrator().CreateTable(&UserIpLimit{})
	}

	if !db.Migrator().HasTable(&RegisterAddFriend{}) {
		fmt.Println("CreateTable RegisterAddFriend")
		db.Migrator().CreateTable(&RegisterAddFriend{})
	}
	if !db.Migrator().HasTable(&Invitation{}) {
		fmt.Println("CreateTable Invitation")
		db.Migrator().CreateTable(&Invitation{})
	}

	if !db.Migrator().HasTable(&ClientInitConfig{}) {
		fmt.Println("CreateTable ClientInitConfig")
		db.Migrator().CreateTable(&ClientInitConfig{})
	}

	if !db.Migrator().HasTable(&UserIpRecord{}) {
		fmt.Println("CreateTable Friend")
		db.Migrator().CreateTable(&UserIpRecord{})
	}

	DB.MysqlDB.db = db
	return
}

func (m *mysqlDB) DefaultGormDB() *gorm.DB {
	return DB.MysqlDB.db
}
