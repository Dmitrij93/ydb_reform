# ydb_reform
In developing...

1) create a file (for example, person.go) and fill it with a sample

```
//ydb_reform
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

2) Build ydb_reform.exe

3) Move this installer to path with structs

4) Run ydb_reform and get buzz
