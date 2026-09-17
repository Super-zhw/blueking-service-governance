#!/bin/bash
# utils.sh - taf 版公共变量和函数定义，其他脚本通过 source 引用

# === 路径变量 ===
# 环境变量中的 APP 名称（可能与实际文件名大小写不一致）
SERVER_NAME_ENV="${BKMS_APP_NAME}"

readonly BKMS_DEV_MODE_PATH="/data/bkms/dev-mode/taf"
readonly BKMS_DEV_MODE_BIN_PATH="${BKMS_DEV_MODE_PATH}/bin"
# 脚本可写副本目录
readonly BKMS_MONITOR_PATH="${BKMS_DEV_MODE_PATH}/scripts"
readonly BKMS_MONITOR_LOGS_PATH="${BKMS_DEV_MODE_PATH}/logs"

readonly PID_CONF="${BKMS_MONITOR_PATH}/pid.conf"
# 通知 monitor 暂停自动拉起
readonly STOP_FLAG_FILE="${BKMS_DEV_MODE_PATH}/scripts/.stop_flag"

# taf 不需要提前查找 bin 路径，进程拉起后动态确定
# 实际二进制目录（由 get_taf_server_info 设置）
TAF_BIN_PATH=""
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

# 在指定目录中查找匹配的二进制文件（不区分大小写）
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

    # 某些环境中文件末尾携带星号，匹配时需要去掉
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

# 从运行中的进程动态获取 taf 服务的二进制路径和名称
# taf 应用通过 taf-start.sh wrap 方式拉起，无法提前知道 bin 路径
# 只有进程拉起后，才能通过 SERVER_NAME_ENV 在进程列表中确定实际路径
# 返回: 设置全局变量 TAF_BIN_PATH 和 SERVER_NAME
get_taf_server_info() {
    # 最高优先级：如果设置了 BKMS_DEVMODE_BINARY_PATH 环境变量，直接使用
    if [ -n "${BKMS_DEVMODE_BINARY_PATH}" ]; then
        TAF_BIN_PATH=$(dirname "${BKMS_DEVMODE_BINARY_PATH}")
        SERVER_NAME=$(basename "${BKMS_DEVMODE_BINARY_PATH}")
        log_info "Using BKMS_DEVMODE_BINARY_PATH override: TAF_BIN_PATH=${TAF_BIN_PATH}, SERVER_NAME=${SERVER_NAME}"
        return 0
    fi

    if [ -z "${SERVER_NAME_ENV}" ]; then
        log_error "SERVER_NAME_ENV is empty, cannot get taf server info"
        return 1
    fi

    # 从 ps -ef 中查找匹配 SERVER_NAME_ENV 的进程，提取完整的可执行文件路径
    local proc_line=$(ps -ef | grep -i "${SERVER_NAME_ENV}" | egrep -v ' grep |taf-start.sh|start.sh|monitor.sh|stop.sh|restart.sh' | sort -rnk 2 | head -n 1)

    if [ -z "${proc_line}" ]; then
        log_warn "No running process found for ${SERVER_NAME_ENV}"
        return 1
    fi

    # 提取可执行文件的完整路径（第8列及之后为命令行，取第一个参数作为可执行文件路径）
    local exe_path=$(echo "${proc_line}" | awk '{print $8}')

    if [ -z "${exe_path}" ]; then
        log_warn "Could not extract executable path from process line"
        return 1
    fi

    # 如果 exe_path 是相对路径或不包含 /，尝试使用 /proc/<pid>/exe
    if [[ "${exe_path}" != /* ]]; then
        local pid=$(echo "${proc_line}" | awk '{print $2}')
        if [ -L "/proc/${pid}/exe" ]; then
            exe_path=$(readlink -f "/proc/${pid}/exe" 2>/dev/null)
        fi
    fi

    # 从完整路径中提取目录和文件名
    TAF_BIN_PATH=$(dirname "${exe_path}")
    SERVER_NAME=$(basename "${exe_path}")

    log_info "Taf server info: TAF_BIN_PATH=${TAF_BIN_PATH}, SERVER_NAME=${SERVER_NAME}"
    return 0
}

# === 进程管理函数 ===

# 获取当前运行的进程数量（通过 SERVER_NAME_ENV 匹配）
# 返回: 进程数量
get_process_count() {
    # shellcheck disable=SC2196
    # shellcheck disable=SC2009
    ps -ef | grep -i "${SERVER_NAME_ENV}" | egrep -v ' grep |taf-start.sh|start.sh|monitor.sh|stop.sh|restart.sh' | sort -rnk 2| wc -l
}

# 获取当前运行的进程 PID
# 返回: 进程 PID
get_process_pid() {
    # shellcheck disable=SC2196
    # shellcheck disable=SC2009
    ps -ef | grep -i "${SERVER_NAME_ENV}" | egrep -v ' grep |taf-start.sh|start.sh|monitor.sh|stop.sh|restart.sh' | sort -rnk 2| head -n 1 | awk '{print $2}'
}

# 获取所有匹配的进程 PID
# 返回: 进程 PID 列表（空格分隔）
get_process_pids() {
    # shellcheck disable=SC2196
    # shellcheck disable=SC2009
    ps -ef | grep -i "${SERVER_NAME_ENV}" | egrep -v ' grep |taf-start.sh|start.sh|monitor.sh|stop.sh|restart.sh' | sort -rnk 2| awk '{print $2}'
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
