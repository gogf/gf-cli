package gen

const templateModelInit = `package {TplPackageName}

import "github.com/gogf/gf/database/gdb"

var (
	// ConfigGroup is the configuration group name for this model.
	ConfigGroup = gdb.DEFAULT_GROUP_NAME
)
`

const templateModelContent = `package {TplPackageName}

import (
	"database/sql"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/database/gdb"
)
{TplExtraImports}
// {TplModelName} is the golang structure for table {TplTableName}.
{TplStructDefine}

// {TplModelName}Model is the model of convenient operations for table {TplTableName}.
type {TplModelName}Model struct {
	*gdb.Model
	TableName string
}

var (
	// {TplModelName}TableName is the table name of {TplTableName}.
	{TplModelName}TableName = "{TplTableName}"
)

// Model{TplModelName} creates and returns a new model object for table {TplTableName}.
func Model{TplModelName}() *{TplModelName}Model {
	return &{TplModelName}Model{
		g.DB(ConfigGroup).Table({TplModelName}TableName).Safe(),
		{TplModelName}TableName,
	}
}

// Inserts does "INSERT...INTO..." statement for inserting current object into table.
func (r *{TplModelName}) Insert() (result sql.Result, err error) {
	return Model{TplModelName}().Data(r).Insert()
}

// Replace does "REPLACE...INTO..." statement for inserting current object into table.
// If there's already another same record in the table (it checks using primary key or unique index),
// it deletes it and insert this one.
func (r *{TplModelName}) Replace() (result sql.Result, err error) {
	return Model{TplModelName}().Data(r).Replace()
}

// Save does "INSERT...INTO..." statement for inserting/updating current object into table.
// It updates the record if there's already another same record in the table
// (it checks using primary key or unique index).
func (r *{TplModelName}) Save() (result sql.Result, err error) {
	return Model{TplModelName}().Data(r).Save()
}

// Update does "UPDATE...WHERE..." statement for updating current object from table.
// It updates the record if there's already another same record in the table
// (it checks using primary key or unique index).
func (r *{TplModelName}) Update() (result sql.Result, err error) {
	return Model{TplModelName}().Data(r).Where(gdb.GetWhereConditionOfStruct(r)).Update()
}

// Delete does "DELETE FROM...WHERE..." statement for deleting current object from table.
func (r *{TplModelName}) Delete() (result sql.Result, err error) {
	return Model{TplModelName}().Where(gdb.GetWhereConditionOfStruct(r)).Delete()
}

// Select overwrite the Select method from gdb.Model for model
// as retuning all objects with specified structure.
func (m *{TplModelName}Model) Select() ([]*{TplModelName}, error) {
	array := ([]*{TplModelName})(nil)
	if err := m.Scan(&array); err != nil {
		return nil, err
	}
	return array, nil
}

// First does the same logistics as One method from gdb.Model for model
// as retuning first/one object with specified structure.
func (m *{TplModelName}Model) First() (*{TplModelName}, error) {
	list, err := m.Select()
	if err != nil {
		return nil, err
	}
	if len(list) > 0 {
		return list[0], nil
	}
	return nil, nil
}
`
