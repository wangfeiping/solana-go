CREATE TABLE IF NOT EXISTS etl_solana_logs_queue (
  id                   INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  timestamp            BIGINT UNSIGNED NOT NULL COMMENT '时间戳',
  chain                CHAR(16) NOT NULL DEFAULT 'solana' COMMENT '代币所属链',
  signature            VARCHAR(88) DEFAULT NULL COMMENT '图标地址',
  log                  VARCHAR(1024) DEFAULT NULL COMMENT 'URI地址',
  status               INT UNSIGNED NOT NULL DEFAULT 0,
  version              INT UNSIGNED NOT NULL DEFAULT 0,

  PRIMARY KEY (id),
  UNIQUE INDEX (signature),
  INDEX status_signature (status,signature) COMMENT '通过处理状态查询日志队列信息索引',
)
ENGINE=InnoDB 
DEFAULT CHARSET=utf8mb4 
COLLATE=utf8mb4_0900_ai_ci;

