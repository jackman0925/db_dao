# DB DAO

一个基于 `sqlx` 的通用且易于使用的 Go DAO（数据访问对象）库。

该库旨在轻松集成到任何 Go 项目中，为数据库操作提供简单一致的 API。

## 特性

- 通用的 CRUD 操作方法。
- 用于构建查询的链式 API。
- 分页支持。
- 事务支持。
- 原生 SQL 执行。

## 安装

这是一个内部库。要使用它，请将 `db_dao` 目录复制到您的项目中。

您的项目中还需要 `sqlx`：
```shell
go get github.com/jmoiron/sqlx
```

## 使用方法

### 初始化

首先，创建一个 `*sqlx.DB` 连接对象。然后，创建一个新的 DAO 实例。

```go
import (
    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq" // 或您的数据库驱动
    "your_project/path/to/db_dao"
)

// ...

db, err := sqlx.Connect("postgres", "user=... password=... dbname=... sslmode=disable")
if err != nil {
    log.Fatalln(err)
}

dao := db_dao.NewDAO[User](db)
```

### 模型（Models）

将您的数据库模型定义为 Go 结构体。使用 `db` 标签将结构体字段映射到数据库列。

```go
type User struct {
    ID   int    `db:"id"`
    Name string `db:"name"`
    Age  int    `db:"age"`
}
```

### Get (单条记录)

```go
var user User
err := dao.Get(db_dao.GetEndPoint[User]{
    Model: &user,
    Table: "users",
    Fields: []string{"id", "name", "age"},
    Conditions: map[string]any{
        "id = ": 1,
    },
})
```

### Select (多条记录)

```go
var users []User
err := dao.Select(db_dao.SelectEndPoint[User]{
    Model: &users,
    Table: "users",
    Fields: []string{"id", "name"},
    Conditions: map[string]any{
        "age > ": 25,
    },
    Appends: []string{"ORDER BY id DESC"},
})
```

### Paginate (分页)

```go
var users []User
total, err := dao.Paginate(db_dao.PageEndPoint[User]{
    Model:    &users,
    Table:    "users",
    Fields:   []string{"id", "name"},
    PageNo:   1,
    PageSize: 10,
    SortField: "id",
    SortOrder: "DESC",
    Conditions: map[string]any{
        "age > ": 20,
    },
})
```

### Insert (插入)

```go
affectedRows, err := dao.Insert(db_dao.InsertEndpoint[User]{
    Table: "users",
    Rows: map[string]any{
        "name": "John Doe",
        "age":  30,
    },
})
```

### Batch Insert (批量插入)

```go
affectedRows, err := dao.BatchInsert(db_dao.BatchInsertEndpoint[User]{
    Table: "users",
    Rows: []map[string]any{
        {"name": "Jane Doe", "age": 28},
        {"name": "Peter Pan", "age": 12},
    },
})
```

### Update (更新)

**注意：** 为安全起见，`Update` 操作需要至少一个条件。

```go
affectedRows, err := dao.Update(db_dao.UpdateEndPoint[User]{
    Table: "users",
    Rows: map[string]any{
        "age": 31,
    },
    Conditions: map[string]any{
        "name = ": "John Doe",
    },
})
```

### Delete (删除)

**注意：** 为安全起见，`Delete` 操作需要至少一个条件。

```go
affectedRows, err := dao.Delete(db_dao.DeleteEndPoint[User]{
    Table: "users",
    Conditions: map[string]any{
        "id = ": 1,
    },
})
```

### 事务 (Transactions)

```go
txDao, err := dao.BeginTx(context.Background())
if err != nil {
    // 处理错误
}

// Defer 回滚/提交逻辑
defer func() {
    if p := recover(); p != nil {
        txDao.Rollback()
        panic(p)
    } else if err != nil {
        txDao.Rollback()
    } else {
        err = txDao.Commit()
    }
}()

// 在事务中执行操作
_, err = txDao.Insert(db_dao.InsertEndpoint[User]{
    Table: "users",
    Rows: map[string]any{"name": "Tx User", "age": 50},
})
if err != nil {
    return // 错误将由 defer 处理
}
```

### 原生 SQL

```go
// Exec (用于 INSERT, UPDATE, DELETE)
affected, err := dao.Exec("UPDATE users SET age = ? WHERE id = ?", 32, 1)

// QueryToSlice (用于 SELECT)
var users []User
err = db_dao.QueryToSlice(dao.GetExecutor(), "SELECT * FROM users WHERE age > ?", &users, 30)
```