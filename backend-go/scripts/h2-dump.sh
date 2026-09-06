#!/usr/bin/env bash
# 用法: ./scripts/h2-dump.sh <h2-jar路径> <demo.mv.db所在目录> <输出目录>
# 例: ./scripts/h2-dump.sh ~/.gradle/.../h2-2.x.jar ../backend/data ./dump
set -euo pipefail

H2_JAR="$1"
DB_DIR="$2"      # 含 demo.mv.db 的目录（不含 demo.mv.db 后缀）
OUT_DIR="$3"
mkdir -p "$OUT_DIR"

# H2 连接串与凭据来自原 application.yaml
URL="jdbc:h2:file:${DB_DIR}/demo"
USER="clash-configs"
PASS="password1."

# H2 的 `user` 是保留字（表名以带引号标识符存储为小写），导出时需加双引号。
# 其它表名以未加引号标识符存储为大写，直接写小写即可（H2 自动折叠为大写）。
TABLES=(
  "clash_configs:clash_configs"
  "clash_configs_merge:clash_configs_merge"
  "clash_configs_merge_config:clash_configs_merge_config"
  "user:\"user\""
)

for entry in "${TABLES[@]}"; do
  name="${entry%%:*}"
  ref="${entry#*:}"
  echo ">> dumping ${name}"
  java -cp "$H2_JAR" org.h2.tools.Shell \
    -url "$URL" -user "$USER" -password "$PASS" \
    -sql "CALL CSVWRITE('${OUT_DIR}/${name}.csv', 'SELECT * FROM ${ref}')"
done
echo ">> done"
