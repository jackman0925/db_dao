# DB DAO

一个基于 `sqlx` 的通用且易于使用的 Go DAO（数据访问对象）库。

该库旨在轻松集成到任何 Go 项目中，为数据库操作提供简单、可测试且一致的 API。

## 特性

- 通用的 CRUD 操作方法。
- 用于构建查询的链式 API。
- 分页支持。
- 完整的事务支持（通过接口）。
- 面向接口的设计，易于测试和模拟（Mock）。

## 安装

这是一个内部库。要使用它，请将 `db_dao` 目录复制到您的项目中，或将其作为私有 Go 模块进行管理。

您的项目中还需要 `sqlx`：
```shell
go get github.com/jmoiron/sqlx
```

## 使用方法

### 1. 初始化

首先，创建一个 `*sqlx.DB` 连接对象。然后，使用 `NewDAO` 创建一个 DAO 实例。推荐将其赋值给 `IDAO[T]` 接口类型，以便于测试。

```go
import (
    "github.com/jmoiron/sqlx"
    // 引入 pgx 驱动 (推荐用于 PostgreSQL)
    _ "github.com/jackc/pgx/v5/stdlib" 
    // 或者引入其他驱动，如 sqlite3
    // _ "github.com/mattn/go-sqlite3"
	"github.com/jackman0925/db_dao"
)

// ...

// 使用 "pgx" 作为驱动名称连接 PostgreSQL
db, err := sqlx.Connect("pgx", "postgres://user:password@localhost:5432/dbname")
if err != nil {
    log.Fatalln(err)
}

// 将 DAO 实例赋值给 IDAO 接口，这是最佳实践
// 库会自动处理不同数据库的占位符差异 (如 PG 的 $1, $2)
var userDAO db_dao.IDAO[User] = db_dao.NewDAO[User](db)
```

### 2. 定义模型

使用 `db` 标签将结构体字段映射到数据库列。

```go
type User struct {
    ID   int64  `db:"id"`
    Name string `db:"name"`
    Age  int    `db:"age"`
}
```

### 3. 基本操作

所有基本操作（Get, Select, Insert, Update, Delete, Paginate）都通过 `IDAO` 接口调用。

**获取单条记录 (Get):**
```go
var user User
err := userDAO.Get(context.Background(), db_dao.GetEndPoint[User]{
    Model:      &user,
    Table:      "users",
    Conditions: map[string]any{"id = ": 1},
})
```

**插入记录 (Insert):**
```go
affectedRows, err := userDAO.Insert(context.Background(), db_dao.InsertEndpoint[User]{
    Table: "users",
    Rows:  map[string]any{"name": "John Doe", "age": 30},
})
```

**复杂查询 (OR Conditions):**
```go
// SELECT * FROM users WHERE (age = 30 OR age = 40)
var users []User
err := userDAO.Select(context.Background(), db_dao.SelectEndPoint[User]{
    Model: &users,
    Table: "users",
    Conditions: map[string]any{
        "or_group": db_dao.Or{
            {"age = ": 30},
            {"age = ": 40},
        },
    },
})
```

### 4. 事务 (Transactions)

事务处理是本库的核心功能之一，通过 `BeginTx` 方法可以轻松实现。

```go
func doSomethingInTransaction(userDAO db_dao.IDAO[User]) (err error) {
    // 1. 开始事务，这将返回一个新的、支持事务的 IDAO 实例
    txDAO, err := userDAO.BeginTx(context.Background())
    if err != nil {
        return err
    }

    // 2. 使用 defer 来确保事务最终会被提交或回滚
    defer func() {
        if p := recover(); p != nil {
            txDAO.Rollback() // 发生 panic，回滚
            panic(p)
        } else if err != nil {
            txDAO.Rollback() // 函数返回错误，回滚
        } else {
            err = txDAO.Commit() // 一切正常，提交
        }
    }()

    // 3. 在事务中执行数据库操作
    // 所有 txDAO 上的操作都在同一个事务中
    if _, err = txDAO.Insert(...); err != nil {
        return err
    }

    if _, err = txDAO.Update(...); err != nil {
        return err
    }

    return nil
}
```
