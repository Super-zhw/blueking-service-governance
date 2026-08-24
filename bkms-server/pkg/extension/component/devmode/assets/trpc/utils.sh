#!/bin/bash
# utils.sh - 公共变量和函数定义，其他脚本通过 source 引用

# === 路径变量 ===
# 环境变量中的 APP 名称（可能与实际文件名大小写不一致）
SERVER_NAME_ENV="${BKMS_APP_NAME}"

readonly BKMS_DEV_MODE_PATH="/data/bkms/dev-mode/trpc"
readonly BKMS_DEV_MODE_BIN_PATH="${BKMS_DEV_MODE_PATH}/bin"
# 脚本可写副本目录
readonly BKMS_MONITOR_PATH="${BKMS_DEV_MODE_PATH}/scripts"
readonly BKMS_MONITOR_LOGS_PATH="${BKMS_DEV_MODE_PATH}/logs"

readonly PID_CONF="${BKMS_MONITOR_PATH}/pid.conf"
# 通知 monitor 暂停自动拉起
readonly STOP_FLAG_FILE="${BKMS_DEV_MODE_PATH}/scripts/.stop_flag"

# 二进制搜索路径（按优先级：模板配置 > 备用路径）
TRPC_BIN_PATH_TEMPLATE="{{.BKMS_TRPC_BIN_PATH}}"
readonly TRPC_BIN_SEARCH_PATHS=(
    "${TRPC_BIN_PATH_TEMPLATE}"
    "/usr/local/trpc/bin"
    "/usr/local/trpc/conf"
)

# 实际二进制目录（由 get_actual_server_name 设置）
TRPC_BIN_PATH=""
# 日志目录（由 get_actual_server_name 设置）
TRPC_LOG_PATH=""
# 实际二进制文件名（处理大小写和星号）
SERVER_NAME=""

source /etc/profile 2>/dev/null || true

# === 日志函数（调用前需定义 LOG_FILE 变量）===

# 基础日志函数
# 参数: $1=日志级别, $@=日志消息
# 注意: 调用此函数前需要在脚本中定义 LOG_FILE 变量
log() {
    local level=$1
    shift
    local message="$@"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    if [ -n "${LOG_FILE}" ]; then
        echo "[${timestamp}] [${level}] ${message}" >> "${LOG_FILE}"
    else
        echo "[${timestamp}] [${level}] ${message}"
    fi
}

log_info() {
    log "INFO" "$@"
}

log_success() {
    log "SUCCESS" "$@"
}

log_warn() {
    log "WARN" "$@"
}

log_error() {
    log "ERROR" "$@"
}

log_fatal() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [FATAL] $*"
    exit 1
}

# === 工具函数 ===

# 在指定目录中查找匹配的二进制文件
# 参数: $1=搜索目录
# 返回: 找到的文件名（stdout），未找到返回空
find_binary_in_dir() {
    local search_dir=$1
    local compact_server_name
    local candidate
    local compact_candidate
    local -a candidates

    if [ -z "${search_dir}" ] || [ ! -d "${search_dir}" ]; then
        return
    fi

    mapfile -t candidates < <(ls "${search_dir}/" 2>/dev/null | sed 's/\*$//')

    # 先按原始形态做大小写不敏感的精确匹配。
    for candidate in "${candidates[@]}"; do
        if [ "${candidate,,}" = "${SERVER_NAME_ENV,,}" ]; then
            printf '%s\n' "${candidate}"
            return 0
        fi
    done

    # 再按简洁形态匹配：双方只去掉连字符和下划线，再做大小写不敏感比较。
    compact_server_name=$(printf '%s' "${SERVER_NAME_ENV}" | tr -d '_-')
    for candidate in "${candidates[@]}"; do
        compact_candidate=$(printf '%s' "${candidate}" | tr -d '_-')
        if [ "${compact_candidate,,}" = "${compact_server_name,,}" ]; then
            printf '%s\n' "${candidate}"
            return 0
        fi
    done

    return 1
}

# 获取实际的二进制文件路径和名称
# 优先级：
# 0. 如果设置了 BKMS_DEVMODE_BINARY_PATH 环境变量（完整路径+二进制名称），直接使用
# 1. 首先尝试在模板配置的路径下查找
# 2. 如果找不到，依次尝试备用路径
# 3. 环境变量中的 BKMS_APP_NAME 可能是小写，但实际文件名可能是大小写混合
# 4. ls 命令输出的可执行文件可能带有 * 后缀，需要去掉
# 返回: 设置全局变量 TRPC_BIN_PATH, SERVER_NAME 和 TRPC_LOG_PATH
get_actual_server_name() {
    # 优先使用 BKMS_DEVMODE_BINARY_PATH 环境变量（完整路径+二进制名称）
    if [ -n "${BKMS_DEVMODE_BINARY_PATH}" ]; then
        TRPC_BIN_PATH=$(dirname "${BKMS_DEVMODE_BINARY_PATH}")
        SERVER_NAME=$(basename "${BKMS_DEVMODE_BINARY_PATH}")
        TRPC_LOG_PATH="$(dirname "${TRPC_BIN_PATH}")/log"
        log_info "Using BKMS_DEVMODE_BINARY_PATH: ${BKMS_DEVMODE_BINARY_PATH} (dir=${TRPC_BIN_PATH}, name=${SERVER_NAME})"
        return 0
    fi

    if [ -z "${SERVER_NAME_ENV}" ]; then
        log_error "SERVER_NAME_ENV is empty, cannot get actual server name"
        return 1
    fi
    
    # 依次在各个路径中查找二进制文件
    for search_path in "${TRPC_BIN_SEARCH_PATHS[@]}"; do
        if [ -z "${search_path}" ]; then
            continue
        fi
        
        local found_name=$(find_binary_in_dir "${search_path}")
        
        if [ -n "${found_name}" ]; then
            TRPC_BIN_PATH="${search_path}"
            SERVER_NAME="${found_name}"
            # 设置日志路径（与 TRPC_BIN_PATH 平级，位于其父目录下的 log 目录）
            TRPC_LOG_PATH="$(dirname "${TRPC_BIN_PATH}")/log"
            
            if [ "${search_path}" = "${TRPC_BIN_PATH_TEMPLATE}" ]; then
                log_info "Binary found in configured path: ${TRPC_BIN_PATH}/${SERVER_NAME}"
            else
                log_info "Binary found in fallback path: ${TRPC_BIN_PATH}/${SERVER_NAME} (configured path was: ${TRPC_BIN_PATH_TEMPLATE})"
            fi
            return 0
        fi
    done
    
    # 如果所有路径都找不到，报错退出
    log_fatal "Could not find binary '${SERVER_NAME_ENV}' in any of the search paths: ${TRPC_BIN_SEARCH_PATHS[*]}"
}

# === 进程管理函数 ===

# 获取当前运行的进程数量
# 返回: 进程数量
get_process_count() {
    ps -ef | grep -i ${TRPC_BIN_PATH}/${SERVER_NAME} | grep -v ' grep ' | wc -l
}

# 获取当前运行的进程 PID
# 返回: 进程 PID
get_process_pid() {
    ps -ef | grep -i ${TRPC_BIN_PATH}/${SERVER_NAME} | grep -v ' grep ' | head -n 1 | awk '{print $2}'
}

# 获取所有匹配的进程 PID
# 返回: 进程 PID 列表（空格分隔）
get_process_pids() {
    ps -ef | grep -i ${TRPC_BIN_PATH}/${SERVER_NAME} | grep -v ' grep ' | awk '{print $2}'
}

# 获取记录的 PID
# 返回: 记录的 PID
get_recorded_pid() {
    if [ -f "${PID_CONF}" ]; then
        cat "${PID_CONF}"
    fi
}

# 检查进程是否存在
# 参数: $1=进程PID
# 返回: 0=存在, 1=不存在
is_process_running() {
    local pid=$1
    if [ -z "${pid}" ]; then
        return 1
    fi
    ps -p "${pid}" >/dev/null 2>&1
}

# === 停止标志管理 ===

# 创建停止标志文件
create_stop_flag() {
    touch "${STOP_FLAG_FILE}"
    log_info "Stop flag created: ${STOP_FLAG_FILE}"
}

# 清除停止标志文件
clear_stop_flag() {
    if [ -f "${STOP_FLAG_FILE}" ]; then
        rm -f "${STOP_FLAG_FILE}"
        log_info "Stop flag cleared: ${STOP_FLAG_FILE}"
    fi
}

# 检查停止标志是否存在
# 返回: 0=存在（已停止）, 1=不存在（正常运行）
is_stop_flag_set() {
    [ -f "${STOP_FLAG_FILE}" ]
}
