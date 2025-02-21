#!/bin/sh
# 检查是否以 root 权限运行
if [ "$(id -u)" -ne 0 ]; then
    echo "请以 root 权限运行此脚本（使用 sudo 或切换到 root 用户）"
    exit 1
fi
# 定义变量
REPO_FILE="/etc/apk/repositories"
BACKUP_FILE="/etc/apk/repositories.bak"
MIRROR_URL="https://mirrors.tuna.tsinghua.edu.cn/alpine/v3.14"
# 备份原始文件
if [ -f "$REPO_FILE" ]; then
    cp "$REPO_FILE" "$BACKUP_FILE"
    echo "已备份原始文件到 $BACKUP_FILE"
else
    echo "错误：$REPO_FILE 不存在"
    exit 1
fi
# 替换为国内源
echo "正在更换为清华大学镜像源..."
cat > "$REPO_FILE" << EOL
${MIRROR_URL}/main
${MIRROR_URL}/community
EOL
# 检查是否替换成功
if [ $? -eq 0 ]; then
    echo "镜像源更换成功，新配置如下："
    cat "$REPO_FILE"
else
    echo "错误：无法写入 $REPO_FILE"
    exit 1
fi
# 更新包列表
echo "正在更新包列表..."
apk update
if [ $? -eq 0 ]; then
    echo "包列表更新成功！"
else
    echo "错误：包列表更新失败，请检查网络或镜像源地址"
    exit 1
fi
echo "更换国内源完成！"