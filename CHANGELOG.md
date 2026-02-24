# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v1.0.5] - 2026-02-24

### 修复 (Fixed)

- **[Bug]** `BeginTx` 忽略了 `opts ...*sql.TxOptions` 参数，始终传 `nil` 给底层 `BeginTxx`，现已正确传递用户指定的事务选项。
- **[Bug]** `buildConditions` 中 `IN` 查询生成错误 SQL（双重 `IN IN`），`sqlx.In(" IN (?)", v)` 产生 `" IN (?,?)"` 后再拼接导致 `field IN  IN (?)`。已修复为正确的 `field IN (?, ?)`。
- **[Bug]** `buildConditions` 中当 condition value 为 `nil` 时，`reflect.ValueOf(nil).Kind()` 会 panic。已增加 nil 值保护。
- **[Bug]** `Update` 方法使用 `append(rowsArgs, conditionsArgs...)` 可能在 `rowsArgs` 有剩余容量时污染原 slice 数据。已改用显式新 slice 构造。
- **[Bug]** `PageEndPoint.point2pageSql()` 中 `PageNo <= 0` 或 `PageSize <= 0` 会产生负数 OFFSET 或无效 LIMIT。已增加参数校验。

### 变更 (Changed)

- 移除 `GetEndPoint`、`SelectEndPoint` 和 `UpdateEndPoint` 中未被使用的 `Options` 字段，减少 API 误导。
- `batch_insert.go` 统一使用 `[]any` 替代 `[]interface{}`，与项目其他文件保持一致。

### 新增 (Added)

- 新增 `builder_test.go`：覆盖所有 SQL 构建函数的单元测试，包含 IN 子句、nil 值、空输入等边界场景。
- 新增 `endpoint_test.go`：覆盖所有 endpoint `point2Sql()` 方法的单元测试。
- 新增集成测试：`BatchInsert`、`IN` 查询、`Fields` 选择、`Appends` 排序、分页排序、无效分页参数、空表/空行/空条件错误路径、`BeginTx` 带 `TxOptions`、`Get` 无结果等。
- 测试覆盖率从原有的基础水平提升至 **89.8%**。

## [v1.0.4] - 2026-01-18

### 新增 (Added)

- 引入 `jackc/pgx` 驱动库支持。
- 新增 `examples/pgx_example` 目录，演示如何使用 Postgres 驱动 `pgx`。
- 验证此库与 `pgx` 驱动的兼容性。

## [v1.0.3] - 2025-11-26

### 新增 (Added)

- 搜索查询条件支持 OR 逻辑 (`db_dao.Or` 类型)。

## [v1.0.2] - 2025-09-16

这是一个重要的重构版本，引入了面向接口的设计，极大地提升了库的可测试性和健壮性。

### 新增 (Added)

- `interfaces.go` 文件，定义了核心的 `IDAO[T]` 和 `Executor` 接口。
- 完整的单元测试套件 (`dao_test.go`)，覆盖了所有核心 CRUD、分页和事务功能，确保了代码质量。

### 变更 (Changed)

- **[重大变更]** `DAO` 结构体被重构，现在内部只包含一个 `Executor` 接口，以统一处理数据库连接和事务。
- **[重大变更]** `BeginTx` 方法现在返回 `IDAO[T]` 接口和一个错误，使得事务处理更加清晰和类型安全。
- **[重大变更]** `Commit` 和 `Rollback` 方法现在是 `IDAO` 接口的一部分，并且其实现已被修正，只在事务性 `DAO` 上有效。
- `Paginate` 方法的返回值从 `int32` 修正为 `int64`，以保持一致性。
- 更新了 `examples/example.go` 以展示基于接口和新事务模型的最佳实践。
- 更新了 `README.md` 以反映新的 API 设计和使用方法。

### 修复 (Fixed)

- 修复了 `DAO` 实现与 `IDAO` 接口之间多个不匹配的问题，包括方法签名和返回值类型。
- 修复了事务和非事务操作中不一致的数据库调用逻辑。

## [v1.0.1]
### Added
- Added `CHANGELOG.md` to track project changes.
- Added `example.go` to demonstrate library usage.