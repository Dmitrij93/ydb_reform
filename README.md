# ydb_reform
In developing...

create a file (for example, person.go) and fill it with a sample

`//ydb_reform
package main

import (
	"time"
)

type Person struct {
	ID        int32      `ydb_reform:"id,pk"`
	Name      string     `ydb_reform:"name"`
	Email     *string    `ydb_reform:"email"`
	CreatedAt time.Time  `ydb_reform:"created_at"`
	UpdatedAt *time.Time `ydb_reform:"updated_at"`
}`
