package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

const (
	messageRequired  = " is required"
	emptyStr         = ""
	tagDB            = "db"
	tagRequired      = "required"
	tagRequeridTrue  = "1"
	tagJSON          = "json"
	sqlUpdate        = "UPDATE "
	sqlSet           = " SET "
	sqlBindAjust     = 1
	sqlValues        = "VALUES"
	sqlSeparator     = ","
	sqlBindValue     = ":"
	sqlInsert        = "INSERT INTO "
	parenthesisLeft  = "("
	parenthesisRight = ")"
	tagUpdate        = "update"
	sqlEqual         = "="
)

func main() {
	const table = "TBL_EXEMPLO"
	//Uma strutura de exemplo
	type Exemple struct { //required:"1" = campos requeridos para o insert && update:"1" campos que vão ser updatados
		ID          uint32 `db:"ID" json:"id" required:"1" update:"1"`
		Name        string `db:"NAME" json:"name" required:"1" update:"1"`
		Description string `db:"DESCRIPTION" json:"description" required:"1" update:"1"`
		Status      *bool  `db:"STATUS" json:"status" required:"1" update:"1"`
	}
	//instanciar a struct/classe
	var e Exemple
	str := []byte(`{"id": 1, "name": "Exemplo de um insert", "description": "Descrição do insert", "status": true}`)
	json.Unmarshal(str, &e)
	//gerar o sql dinâmico de insert
	strInsert, argsInsert, err := Insert(&e, table)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(strInsert)     //INSERT INTO TBL_EXEMPLO(ID,NAME,DESCRIPTION,STATUS)VALUES(:0,:1,:2,:3)
	fmt.Println(argsInsert...) //1 Exemplo de um insert Descrição do insert 1

	//gerar sql dinamico de update
	strUpdate, argsUpdate := Update(&e, table)
	fmt.Println(strUpdate + " WHERE ID = 2") // UPDATE TBL_EXEMPLOSET ID=:0,NAME=:1,DESCRIPTION=:2,STATUS=:3 WHERE ID = 2
	fmt.Println(argsUpdate)                  //[1 Exemplo de um insert Descrição do insert 1]

	type Exemple2 struct { //required:"1" = campos requeridos para o insert && update:"1" campos que vão ser updatados
		ID          uint32  `db:"ID" json:"id" required:"1" update:"1"`
		Name        *string `db:"NAME" json:"name" update:"1" required:"1"`
		Description string  `db:"DESCRIPTION" json:"description" required:"1"`
		Status      *bool   `db:"STATUS" json:"status" required:"1" update:"1"`
	}

	var e2 Exemple2
	str2 := []byte(`{"id": 2,"description": "Exemplo de descrição", "status": true}`)
	json.Unmarshal(str2, &e2)
	//gerar o sql dinâmico de insert
	strInsert2, argsInsert2, err2 := Insert(&e2, table)

	//gera o erro pois o campo nome é requerido
	if err2 != nil {
		fmt.Println(err2) //name is required
	}
	fmt.Println(strInsert2) //""

	fmt.Println(argsInsert2) //[]

	//gerar sql dinamico de update
	strUpdate2, argsUpdate2 := Update(&e2, table)
	fmt.Println(strUpdate2 + " WHERE ID = 2") // UPDATE TBL_EXEMPLO SET ID=:0,NAME=:1,STATUS=:2 WHERE ID = 2
	//como no struct, a tag update do campo descrição não estava marcado com a tag "update:"1"" não foi
	//gerado ele na string de update

	fmt.Println(argsUpdate2) //[Exemplo de descrição]
}

//Insert monta a string de insert
func Insert(o interface{}, tableName string) (string, []interface{}, error) {
	params := emptyStr
	values := emptyStr
	var args []interface{}
	val := reflect.ValueOf(o).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		tagField := val.Type().Field(i).Tag
		tagFieldDB := tagField.Get(tagDB)
		if tagFieldDB != emptyStr {
			if valueField.Kind() == reflect.Ptr {
				if valueField.IsNil() {
					if tagField.Get(tagRequired) == tagRequeridTrue {
						return emptyStr, nil, errors.New(tagField.Get(tagJSON) + messageRequired)
					}
				}
				args = append(args, reflectValueToTypePrimitive(valueField.Elem()))
			} else {
				if !valueField.IsValid() {

					if tagField.Get(tagRequired) == tagRequeridTrue {
						return emptyStr, nil, errors.New(tagField.Get(tagJSON) + messageRequired)
					}
				}
				args = append(args, reflectValueToTypePrimitive(valueField))
			}
			size := strconv.Itoa(len(args) - sqlBindAjust)

			if params == emptyStr {
				params = tagFieldDB
				values = sqlBindValue + size
			} else {
				params += sqlSeparator + tagFieldDB
				values += sqlSeparator + sqlBindValue + size
			}
		}
	}
	sqlInsert := sqlInsert + tableName + concatenationParenthesis(params) + sqlValues + concatenationParenthesis(values)
	return sqlInsert, args, nil
}

func boolToInt(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func reflectValueToTypePrimitive(value reflect.Value) interface{} {
	switch value.Kind() {
	case reflect.Uint32, reflect.Uint8, reflect.Uint16, reflect.Uint:
		return value.Uint()
	case reflect.Bool:
		return boolToInt(value.Bool())
	case reflect.String:
		return value.String()
	case reflect.Float64, reflect.Float32:
		return value.Float()
	}
	return value.Int()
}

func concatenationParenthesis(text string) string {
	return parenthesisLeft + text + parenthesisRight
}

//Update : Generate a sql of Update
func Update(o interface{}, tableName string) (string, []interface{}) {
	query := emptyStr
	var args []interface{}
	val := reflect.ValueOf(o).Elem()

	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		tagField := val.Type().Field(i).Tag
		tagFieldDB := tagField.Get(tagDB)

		if tagFieldDB != emptyStr {
			if tagField.Get(tagUpdate) == emptyStr {
				if valueField.Kind() == reflect.Ptr {
					args = append(args, reflectValueToTypePrimitive(valueField.Elem()))
				} else {
					args = append(args, reflectValueToTypePrimitive(valueField))
				}
				size := strconv.Itoa(len(args) - sqlBindAjust)

				if query == emptyStr {
					query = tagFieldDB + sqlEqual + sqlBindValue + size
				} else {
					query += sqlSeparator + tagFieldDB + sqlEqual + sqlBindValue + size
				}
			}
		}
	}
	return sqlUpdate + tableName + sqlSet + query, args
}
