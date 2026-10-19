# Requirements Document

## Introduction

在订阅管理页，用户当前只能修改单个订阅的 cron 表达式和启停状态。当订阅源链接变化、需要调整分区/选集、下载类型或画质时，用户无法直接修改，只能删除后重新添加。本功能为已订阅项目提供"重新配置"能力：允许用户基于已有订阅重新解析链接、选择分区、设置下载参数并保存。

## Glossary

- **订阅（Subscription）**：用户添加的订阅源，类型包括合集（season）、系列（series）、收藏夹（fav）、UP 空间（space）、番剧（bangumi）。
- **重新配置（Reconfigure）**：对已有订阅重新执行链接解析、分区选择、下载类型/画质/cron 设置并保存的过程。
- **分区（Section）**：订阅源下的稿件分组，例如番剧的正片/特别篇、合集的各分区。
- **订阅源信息**：订阅的类型、资源标识（resourceId）、标题、封面和 UP 主（owner）。

## Requirements

### Requirement 1: 提供重新配置入口

**User Story:** AS 订阅用户, I want 在订阅卡片上看到"重新配置"按钮, so that 我可以修改已有订阅的配置。

#### Acceptance Criteria

1. WHEN 订阅列表加载完成，系统 SHALL 在每个订阅卡片上显示"重新配置"按钮。
2. WHEN 用户点击"重新配置"按钮，系统 SHALL 展示该订阅的重新配置表单。
3. WHEN 用户取消或关闭重新配置表单，系统 SHALL 不修改该订阅的任何数据。

### Requirement 2: 重新配置表单回填现有配置

**User Story:** AS 订阅用户, I want 重新配置表单回填当前订阅的链接、下载类型、画质、cron 和已选分区, so that 我可以在此基础上修改。

#### Acceptance Criteria

1. WHEN 重新配置表单打开，系统 SHALL 使用该订阅当前的链接、下载类型、画质、cron 和已选分区填充表单。
2. WHEN 表单回填完成，系统 SHALL 默认选中该订阅当前已启用的分区。

### Requirement 3: 加载最新分区列表

**User Story:** AS 订阅用户, I want 重新配置弹窗展示该订阅最新的分区列表, so that 我可以调整分区/选集。

#### Acceptance Criteria

1. WHEN 重新配置弹窗打开，系统 SHALL 使用该订阅的原始链接重新解析并展示最新订阅标题、类型、总条数和分区列表。
2. WHEN 分区列表加载完成，系统 SHALL 默认选中该订阅当前已启用的分区。
3. IF 重新解析失败，系统 SHALL 使用该订阅当前已保存的分区列表作为回退展示，并提示解析失败原因。

### Requirement 4: 保存重新配置

**User Story:** AS 订阅用户, I want 保存重新配置后的订阅参数, so that 订阅按新配置运行。

#### Acceptance Criteria

1. WHEN 用户点击"保存"且至少选择了一个分区，系统 SHALL 调用更新接口保存新的分区、下载类型、画质和 cron。
2. WHEN 保存成功，系统 SHALL 刷新订阅列表并关闭重新配置弹窗。
3. IF cron 表达式格式错误，系统 SHALL 提示错误信息且不保存任何修改。
4. IF 保存过程中发生网络错误，系统 SHALL 展示错误信息并保持表单内容不变。

### Requirement 5: 取消与数据完整性

**User Story:** AS 订阅用户, I want 在重新配置过程中可以安全取消且不影响已有数据, so that 避免误操作。

#### Acceptance Criteria

1. WHEN 用户点击"取消"，系统 SHALL 关闭重新配置弹窗且不修改订阅数据。
2. WHEN 保存成功后，系统 SHALL 保留订阅的历史检查记录和已记录稿件条目。
3. WHEN 保存成功后，系统 SHALL 保持订阅的链接和订阅源信息不变。