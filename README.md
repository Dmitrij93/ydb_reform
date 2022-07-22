# ydb_reform
In developing...

create a file (for example, person.go) and fill it with a sample

```//ydb_reform
package main

import (
	"time"
)

type Person struct {
	ID        int32      `ydb:"id,primary"`
	Name      string     `ydb:"name"`
	Email     *string    `ydb:"email"`
	CreatedAt time.Time  `ydb:"created_at"`
	UpdatedAt *time.Time `ydb:"updated_at"`
}
```
