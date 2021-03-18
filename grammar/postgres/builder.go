package postgres

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/yaoapp/xun/dbal"
	"github.com/yaoapp/xun/utils"
)

// SQLAddColumn return the add column sql for table create
func (grammarSQL Postgres) SQLAddColumn(db *sqlx.DB, column *dbal.Column) string {
	types := grammarSQL.Types
	quoter := grammarSQL.Quoter

	// `id` bigint(20) unsigned NOT NULL,
	typ, has := types[column.Type]
	if !has {
		typ = "VARCHAR"
	}
	if column.Precision != nil && column.Scale != nil && (typ == "NUMBERIC" || typ == "DECIMAL") {
		typ = fmt.Sprintf("%s(%d,%d)", typ, utils.IntVal(column.Precision), utils.IntVal(column.Scale))
	} else if strings.Contains(typ, "TIMESTAMP(%d)") || strings.Contains(typ, "TIME(%d)") {
		DateTimePrecision := utils.IntVal(column.DateTimePrecision, 0)
		typ = fmt.Sprintf(typ, DateTimePrecision)
	} else if typ == "BYTEA" {
		typ = "BYTEA"
	} else if typ == "ENUM" {
		typ = strings.ToLower("ENUM__" + strings.Join(column.Option, "_EOPT_"))
	} else if column.Length != nil {
		typ = fmt.Sprintf("%s(%d)", typ, utils.IntVal(column.Length))
	}

	unsigned := ""
	nullable := utils.GetIF(column.Nullable, "NULL", "NOT NULL").(string)
	defaultValue := utils.GetIF(column.Default != nil, fmt.Sprintf("DEFAULT %v", column.Default), "").(string)
	// comment := utils.GetIF(utils.StringVal(column.Comment) != "", fmt.Sprintf("COMMENT %s", quoter.VAL(column.Comment, db)), "").(string)
	collation := utils.GetIF(utils.StringVal(column.Collation) != "", fmt.Sprintf("COLLATE %s", utils.StringVal(column.Collation)), "").(string)
	extra := ""
	if utils.StringVal(column.Extra) != "" {
		if typ == "BIGINT" {
			typ = "BIGSERIAL"
		} else if typ == "SMALLINT" {
			typ = "SMALLSERIAL"
		} else {
			typ = "SERIAL"
		}
		nullable = ""
		defaultValue = ""
	}

	if typ == "IPADDRESS" { // ipAddress
		typ = "integer"
	}

	sql := fmt.Sprintf(
		"%s %s %s %s %s %s %s",
		quoter.ID(column.Name, db), typ, unsigned, nullable, defaultValue, extra, collation)

	sql = strings.Trim(sql, " ")
	return sql
}

// SQLAddComment return the add comment sql for table create
func (grammarSQL Postgres) SQLAddComment(db *sqlx.DB, column *dbal.Column) string {
	comment := utils.GetIF(
		utils.StringVal(column.Comment) != "",
		fmt.Sprintf(
			"COMMENT on column %s.%s is %s;",
			grammarSQL.ID(column.TableName, db),
			grammarSQL.ID(column.Name, db),
			grammarSQL.VAL(column.Comment, db),
		), "").(string)

	if column.Type == "ipAddress" { // ipAddress
		comment = fmt.Sprintf("COMMENT on column %s.%s is %s;",
			grammarSQL.ID(column.TableName, db),
			grammarSQL.ID(column.Name, db),
			grammarSQL.VAL(fmt.Sprintf("T:%s|%s", column.Type, utils.StringVal(column.Comment)), db),
		)
	}
	return comment
}

// SQLAddIndex  return the add index sql for table create
func (grammarSQL Postgres) SQLAddIndex(db *sqlx.DB, index *dbal.Index) string {
	quoter := grammarSQL.Quoter
	indexTypes := grammarSQL.IndexTypes
	typ, has := indexTypes[index.Type]
	if !has {
		typ = "KEY"
	}

	// UNIQUE KEY `unionid` (`unionid`) COMMENT 'xxxx'
	// IS JSON
	columns := []string{}
	isJSON := false
	for _, column := range index.Columns {
		columns = append(columns, quoter.ID(column.Name, db))
		if column.Type == "json" || column.Type == "jsonb" {
			isJSON = true
		}
	}
	if isJSON {
		return ""
	}

	comment := ""
	if index.Comment != nil {
		comment = fmt.Sprintf("COMMENT %s", quoter.VAL(index.Comment, db))
	}
	name := quoter.ID(index.Name, db)

	sql := ""
	if typ == "PRIMARY KEY" {
		sql = fmt.Sprintf(
			"%s (%s) %s",
			typ, strings.Join(columns, ","), comment)
	} else {
		sql = fmt.Sprintf(
			"CREATE %s %s ON %s (%s)",
			typ, name, quoter.ID(index.TableName, db), strings.Join(columns, ","))
	}
	return sql
}

// SQLAddPrimary return the add primary key sql for table create
func (grammarSQL Postgres) SQLAddPrimary(db *sqlx.DB, primary *dbal.Primary) string {

	quoter := grammarSQL.Quoter

	// PRIMARY KEY `unionid` (`unionid`) COMMENT 'xxxx'
	columns := []string{}
	for _, column := range primary.Columns {
		columns = append(columns, quoter.ID(column.Name, db))
	}

	sql := fmt.Sprintf(
		"PRIMARY KEY (%s)",
		strings.Join(columns, ","))

	return sql
}
