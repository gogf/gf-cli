package consts

const TemplateGenDaoDoContent = `
// =================================================================================
// Code generated by GoFrame CLI tool. DO NOT EDIT. Created at {TplDatetime}
// =================================================================================

package do

{TplPackageImports}

// {TplTableNameCamelCase} is the golang structure of table {TplTableName} for DAO operations like Where/Data.
{TplStructDefine}
`
