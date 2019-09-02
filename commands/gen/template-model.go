package gen

const templateModel = `
// {Template}Model is the model of convenient operations for table {template}.
type {Template}Model struct {
	*gdb.Model
	TableName string
}

var (
	// {Template}TableName is the table name of {template}.
	{Template}TableName = "{template}"
)

// Model{Template} creates and returns a new model object for table {template}.
func Model{Template}() *{Template}Model {
	return &{Template}Model{
		g.DB(ConfigGroup).Table({Template}TableName).Safe(),
		{Template}TableName,
	}
}

// Inserts does "INSERT...INTO..." statement for inserting current object into table.
func (r *{Template}) Insert() (result sql.Result, err error) {
	return Model{Template}().Data(r).Insert()
}

// Replace does "REPLACE...INTO..." statement for inserting current object into table.
// If there's already another same record in the table (it checks using primary key or unique index),
// it deletes it and insert this one.
func (r *{Template}) Replace() (result sql.Result, err error) {
	return Model{Template}().Data(r).Replace()
}

// Save does "INSERT...INTO..." statement for inserting/updating current object into table.
// It updates the record if there's already another same record in the table
// (it checks using primary key or unique index).
func (r *{Template}) Save() (result sql.Result, err error) {
	return Model{Template}().Data(r).Save()
}

// Update does "UPDATE...WHERE..." statement for updating current object from table.
// It updates the record if there's already another same record in the table
// (it checks using primary key or unique index).
func (r *{Template}) Update() (result sql.Result, err error) {
	return Model{Template}().Data(r).Where(gdb.GetWhereConditionOfStruct(r)).Update()
}

// Delete does "DELETE FROM...WHERE..." statement for deleting current object from table.
func (r *{Template}) Delete() (result sql.Result, err error) {
	return Model{Template}().Where(gdb.GetWhereConditionOfStruct(r)).Delete()
}

// Select overwrite the Select method from gdb.Model for model
// as retuning all objects with specified structure.
func (m *{Template}Model) Select() ([]*{Template}, error) {
	array := ([]*{Template})(nil)
	if err := m.Scan(&array); err != nil {
		return nil, err
	}
	return array, nil
}

// First does the same logistics as One method from gdb.Model for model
// as retuning first/one object with specified structure.
func (m *{Template}Model) First() (*{Template}, error) {
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
