# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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