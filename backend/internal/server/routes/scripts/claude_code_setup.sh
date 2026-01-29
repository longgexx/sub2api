#!/bin/bash
#
# Claude Code Setup Script
# Claude Code 一键配置脚本
# 此脚本由服务端动态生成，已包含配置信息
#

set -e

# ============================================================
# 服务端配置（由服务端动态生成）
# ============================================================
SITE_NAME="{{SITE_NAME}}"
API_BASE_URL="{{API_BASE_URL}}"
SERVER_URL="{{SERVER_URL}}"

# ============================================================
# Colors / 颜色定义
# ============================================================
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# Language (default: zh = Chinese)
LANG_CHOICE="zh"

# ============================================================
# Language functions / 语言函数
# ============================================================

msg() {
    local key="$1"
    if [ "$LANG_CHOICE" = "en" ]; then
        msg_en "$key"
    else
        msg_zh "$key"
    fi
}

msg_zh() {
    case "$1" in
        "info") echo "信息" ;;
        "success") echo "成功" ;;
        "warning") echo "警告" ;;
        "error") echo "错误" ;;
        "select_lang") echo "请选择语言 / Select language" ;;
        "lang_zh") echo "中文" ;;
        "lang_en") echo "English" ;;
        "enter_choice") echo "请输入选择 (默认: 1)" ;;
        "title") echo "Claude Code 配置脚本 for $SITE_NAME" ;;
        "api_base_url") echo "API Base URL" ;;
        "api_key_prompt") echo "请输入您的 API Key" ;;
        "api_key_hint") echo "请先在 $SITE_NAME Web 界面获取 API Key" ;;
        "api_key_empty") echo "API Key 不能为空" ;;
        "env_check_title") echo "环境冲突检测" ;;
        "checking_env") echo "检测 Shell 环境变量..." ;;
        "checking_shell") echo "检测 Shell 配置文件..." ;;
        "checking_claude") echo "检测 Claude 配置文件..." ;;
        "checking_system") echo "检测系统级配置..." ;;
        "found_env") echo "发现环境变量" ;;
        "found_file") echo "发现配置文件" ;;
        "contains_anthropic") echo "包含 ANTHROPIC 配置" ;;
        "no_conflicts") echo "未发现环境冲突" ;;
        "conflicts_found") echo "发现潜在冲突配置" ;;
        "cleanup_prompt") echo "是否清理冲突配置？" ;;
        "cleanup_title") echo "清理冲突配置" ;;
        "backup_dir") echo "备份目录" ;;
        "backed_up") echo "已备份" ;;
        "cleaned") echo "已清理" ;;
        "cleanup_done") echo "清理完成" ;;
        "skip_cleanup") echo "跳过清理，可能导致配置冲突" ;;
        "config_title") echo "配置 Claude Code" ;;
        "created_config") echo "已创建 Claude 配置" ;;
        "added_shell_env") echo "已添加 Shell 环境变量配置" ;;
        "verify_title") echo "验证配置" ;;
        "testing_api") echo "测试 API 连接..." ;;
        "api_test_pass") echo "API 连接测试通过" ;;
        "api_test_warn") echo "API 连接测试返回" ;;
        "api_test_hint") echo "这可能是网络问题或服务器暂时不可用" ;;
        "verifying_files") echo "验证配置文件..." ;;
        "config_verified") echo "Claude 配置文件已正确设置" ;;
        "config_missing") echo "Claude 配置文件缺少 ANTHROPIC_BASE_URL" ;;
        "config_not_exist") echo "Claude 配置文件不存在" ;;
        "complete_title") echo "配置完成" ;;
        "next_steps") echo "下一步操作" ;;
        "step1") echo "重新打开终端或运行" ;;
        "step2") echo "运行 'claude' 命令开始使用" ;;
        "help_hint") echo "如遇问题，请访问" ;;
        "yes_no") echo "(Y/n)" ;;
        "current_config") echo "当前配置" ;;
        "config_active") echo "配置已在当前终端生效" ;;
        *) echo "$1" ;;
    esac
}

msg_en() {
    case "$1" in
        "info") echo "INFO" ;;
        "success") echo "SUCCESS" ;;
        "warning") echo "WARNING" ;;
        "error") echo "ERROR" ;;
        "select_lang") echo "Select language / 请选择语言" ;;
        "lang_zh") echo "中文" ;;
        "lang_en") echo "English" ;;
        "enter_choice") echo "Enter your choice (default: 1)" ;;
        "title") echo "Claude Code Setup Script for $SITE_NAME" ;;
        "api_base_url") echo "API Base URL" ;;
        "api_key_prompt") echo "Enter your API Key" ;;
        "api_key_hint") echo "Please get your API Key from $SITE_NAME web interface first" ;;
        "api_key_empty") echo "API Key cannot be empty" ;;
        "env_check_title") echo "Environment Conflict Check" ;;
        "checking_env") echo "Checking Shell environment variables..." ;;
        "checking_shell") echo "Checking Shell configuration files..." ;;
        "checking_claude") echo "Checking Claude configuration files..." ;;
        "checking_system") echo "Checking system-level configuration..." ;;
        "found_env") echo "Found environment variable" ;;
        "found_file") echo "Found configuration file" ;;
        "contains_anthropic") echo "contains ANTHROPIC config" ;;
        "no_conflicts") echo "No environment conflicts found" ;;
        "conflicts_found") echo "Potential conflicts found" ;;
        "cleanup_prompt") echo "Clean up conflicting configurations?" ;;
        "cleanup_title") echo "Cleaning Up Conflicts" ;;
        "backup_dir") echo "Backup directory" ;;
        "backed_up") echo "Backed up" ;;
        "cleaned") echo "Cleaned" ;;
        "cleanup_done") echo "Cleanup complete" ;;
        "skip_cleanup") echo "Skipping cleanup, may cause configuration conflicts" ;;
        "config_title") echo "Configuring Claude Code" ;;
        "created_config") echo "Created Claude configuration" ;;
        "added_shell_env") echo "Added Shell environment variable configuration" ;;
        "verify_title") echo "Verifying Configuration" ;;
        "testing_api") echo "Testing API connection..." ;;
        "api_test_pass") echo "API connection test passed" ;;
        "api_test_warn") echo "API connection test returned" ;;
        "api_test_hint") echo "This may be a network issue or server temporarily unavailable" ;;
        "verifying_files") echo "Verifying configuration files..." ;;
        "config_verified") echo "Claude configuration file correctly set" ;;
        "config_missing") echo "Claude configuration file missing ANTHROPIC_BASE_URL" ;;
        "config_not_exist") echo "Claude configuration file does not exist" ;;
        "complete_title") echo "Configuration Complete" ;;
        "next_steps") echo "Next Steps" ;;
        "step1") echo "Reopen terminal or run" ;;
        "step2") echo "Run 'claude' command to start using" ;;
        "help_hint") echo "For issues, please visit" ;;
        "yes_no") echo "(Y/n)" ;;
        "current_config") echo "Current Configuration" ;;
        "config_active") echo "Configuration is now active in current terminal" ;;
        *) echo "$1" ;;
    esac
}

# ============================================================
# Helper functions / 工具函数
# ============================================================

print_info() {
    echo -e "${CYAN}[$(msg info)]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[$(msg success)]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[$(msg warning)]${NC} $1"
}

print_error() {
    echo -e "${RED}[$(msg error)]${NC} $1"
}

print_section() {
    echo ""
    echo -e "${BOLD}${BLUE}=============================================${NC}"
    echo -e "${BOLD}${BLUE}  $1${NC}"
    echo -e "${BOLD}${BLUE}=============================================${NC}"
    echo ""
}

confirm() {
    local prompt="$1"
    local response
    read -p "$prompt $(msg yes_no): " response < /dev/tty
    case "$response" in
        [nN][oO]|[nN])
            return 1
            ;;
        *)
            return 0
            ;;
    esac
}

# ============================================================
# Language selection / 语言选择
# ============================================================

select_language() {
    echo ""
    echo -e "${BOLD}$(msg select_lang)${NC}"
    echo "  1) $(msg lang_zh)"
    echo "  2) $(msg lang_en)"
    echo ""
    read -p "$(msg enter_choice): " choice < /dev/tty
    case "$choice" in
        2)
            LANG_CHOICE="en"
            ;;
        *)
            LANG_CHOICE="zh"
            ;;
    esac
}

# ============================================================
# Banner / 欢迎信息
# ============================================================

show_banner() {
    echo ""
    echo -e "${CYAN}=============================================${NC}"
    echo -e "${CYAN}  $(msg title)${NC}"
    echo -e "${CYAN}=============================================${NC}"
    echo ""
}

# ============================================================
# Environment conflict check / 环境冲突检测
# ============================================================

check_environment_conflicts() {
    local conflicts_found=false
    CONFLICT_COUNT=0

    print_section "$(msg env_check_title)"

    # --- 1. Shell environment variables ---
    print_info "$(msg checking_env)"

    if [ -n "$ANTHROPIC_BASE_URL" ]; then
        CONFLICT_COUNT=$((CONFLICT_COUNT + 1))
        conflicts_found=true
        print_warning "$(msg found_env): ANTHROPIC_BASE_URL=$ANTHROPIC_BASE_URL"
    fi

    if [ -n "$ANTHROPIC_AUTH_TOKEN" ]; then
        CONFLICT_COUNT=$((CONFLICT_COUNT + 1))
        conflicts_found=true
        print_warning "$(msg found_env): ANTHROPIC_AUTH_TOKEN (set)"
    fi

    if [ -n "$ANTHROPIC_API_KEY" ]; then
        CONFLICT_COUNT=$((CONFLICT_COUNT + 1))
        conflicts_found=true
        print_warning "$(msg found_env): ANTHROPIC_API_KEY (set)"
    fi

    # --- 2. Shell configuration files ---
    print_info "$(msg checking_shell)"

    for config_file in "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.bash_profile" "$HOME/.profile" "$HOME/.zprofile"; do
        if [ -f "$config_file" ]; then
            if grep -q "ANTHROPIC_BASE_URL\|ANTHROPIC_AUTH_TOKEN\|ANTHROPIC_API_KEY" "$config_file" 2>/dev/null; then
                CONFLICT_COUNT=$((CONFLICT_COUNT + 1))
                conflicts_found=true
                print_warning "$(msg found_file): $config_file $(msg contains_anthropic)"

                # Show specific lines
                grep -n "ANTHROPIC" "$config_file" 2>/dev/null | head -5 | while read line; do
                    echo -e "    ${YELLOW}-> $line${NC}"
                done
            fi
        fi
    done

    # --- 3. Claude configuration files ---
    print_info "$(msg checking_claude)"

    local claude_config="$HOME/.claude/settings.json"
    if [ -f "$claude_config" ]; then
        if grep -q "ANTHROPIC" "$claude_config" 2>/dev/null; then
            CONFLICT_COUNT=$((CONFLICT_COUNT + 1))
            conflicts_found=true
            print_warning "$(msg found_file): $claude_config $(msg contains_anthropic)"
        fi
    fi

    # --- 4. System-level configuration (read-only warning) ---
    print_info "$(msg checking_system)"

    local managed_config="/etc/claude-code/managed-settings.json"
    if [ -f "$managed_config" ]; then
        print_warning "$(msg found_file): $managed_config (enterprise managed)"
    fi

    # --- 5. macOS launchd environment ---
    if [ "$(uname)" = "Darwin" ]; then
        local launchctl_env=$(launchctl getenv ANTHROPIC_BASE_URL 2>/dev/null || true)
        if [ -n "$launchctl_env" ]; then
            CONFLICT_COUNT=$((CONFLICT_COUNT + 1))
            conflicts_found=true
            print_warning "$(msg found_env) (launchd): ANTHROPIC_BASE_URL"
        fi
    fi

    echo ""

    if [ "$conflicts_found" = true ]; then
        print_warning "$(msg conflicts_found): $CONFLICT_COUNT"
        return 1
    else
        print_success "$(msg no_conflicts)"
        return 0
    fi
}

# ============================================================
# Cleanup conflicts / 清理冲突配置
# ============================================================

cleanup_conflicts() {
    local backup_dir="$HOME/.claude-code-backup-$(date +%Y%m%d%H%M%S)"

    print_section "$(msg cleanup_title)"

    mkdir -p "$backup_dir"
    print_info "$(msg backup_dir): $backup_dir"

    # --- Clean Shell configuration files ---
    for config_file in "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.bash_profile" "$HOME/.profile" "$HOME/.zprofile"; do
        if [ -f "$config_file" ]; then
            if grep -q "ANTHROPIC_BASE_URL\|ANTHROPIC_AUTH_TOKEN\|ANTHROPIC_API_KEY" "$config_file" 2>/dev/null; then
                # Backup
                cp "$config_file" "$backup_dir/$(basename $config_file)"
                print_info "$(msg backed_up): $config_file"

                # Comment out instead of deleting (cross-platform sed)
                if [ "$(uname)" = "Darwin" ]; then
                    sed -i '' 's/^\([^#]*export ANTHROPIC_BASE_URL=.*\)$/# [Disabled] \1/' "$config_file"
                    sed -i '' 's/^\([^#]*export ANTHROPIC_AUTH_TOKEN=.*\)$/# [Disabled] \1/' "$config_file"
                    sed -i '' 's/^\([^#]*export ANTHROPIC_API_KEY=.*\)$/# [Disabled] \1/' "$config_file"
                else
                    sed -i 's/^\([^#]*export ANTHROPIC_BASE_URL=.*\)$/# [Disabled] \1/' "$config_file"
                    sed -i 's/^\([^#]*export ANTHROPIC_AUTH_TOKEN=.*\)$/# [Disabled] \1/' "$config_file"
                    sed -i 's/^\([^#]*export ANTHROPIC_API_KEY=.*\)$/# [Disabled] \1/' "$config_file"
                fi

                print_success "$(msg cleaned): $config_file"
            fi
        fi
    done

    # --- Backup Claude settings.json ---
    local claude_config="$HOME/.claude/settings.json"
    if [ -f "$claude_config" ]; then
        cp "$claude_config" "$backup_dir/settings.json"
        print_info "$(msg backed_up): $claude_config"
    fi

    echo ""
    print_success "$(msg cleanup_done), $(msg backup_dir): $backup_dir"
}

# ============================================================
# Setup Claude Code configuration / 设置 Claude Code 配置
# ============================================================

setup_claude_code() {
    local api_key="$1"

    print_section "$(msg config_title)"

    # --- 1. Create Claude config directory ---
    mkdir -p "$HOME/.claude"

    # --- 2. Generate settings.json ---
    local claude_config="$HOME/.claude/settings.json"

    # Check if config exists and try to merge
    if [ -f "$claude_config" ] && command -v jq &> /dev/null; then
        # Use jq to merge
        local temp_config=$(mktemp)
        if jq --arg base_url "$API_BASE_URL" \
              --arg auth_token "$api_key" \
              '.env.ANTHROPIC_BASE_URL = $base_url | .env.ANTHROPIC_AUTH_TOKEN = $auth_token | .env.CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC = "1"' \
              "$claude_config" > "$temp_config" 2>/dev/null; then
            mv "$temp_config" "$claude_config"
        else
            rm -f "$temp_config"
            create_new_settings "$claude_config" "$api_key"
        fi
    else
        create_new_settings "$claude_config" "$api_key"
    fi

    print_success "$(msg created_config): $claude_config"

    # --- 3. Setup Shell environment variables ---
    setup_shell_env "$api_key"
}

create_new_settings() {
    local config_file="$1"
    local auth_token="$2"

    cat > "$config_file" << EOF
{
  "env": {
    "ANTHROPIC_BASE_URL": "$API_BASE_URL",
    "ANTHROPIC_AUTH_TOKEN": "$auth_token",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1"
  }
}
EOF
}

setup_shell_env() {
    local api_key="$1"

    # Detect user's shell
    local shell_name=$(basename "$SHELL")
    local shell_config=""

    case "$shell_name" in
        bash)
            shell_config="$HOME/.bashrc"
            ;;
        zsh)
            shell_config="$HOME/.zshrc"
            ;;
        *)
            shell_config="$HOME/.profile"
            ;;
    esac

    # Check if our config block already exists and remove it
    if grep -q "$SITE_NAME Claude Code Configuration" "$shell_config" 2>/dev/null; then
        if [ "$(uname)" = "Darwin" ]; then
            sed -i '' "/# === $SITE_NAME Claude Code Configuration ===/,/# === End $SITE_NAME Configuration ===/d" "$shell_config"
        else
            sed -i "/# === $SITE_NAME Claude Code Configuration ===/,/# === End $SITE_NAME Configuration ===/d" "$shell_config"
        fi
    fi

    # Add configuration
    cat >> "$shell_config" << EOF

# === $SITE_NAME Claude Code Configuration ===
# Added by $SITE_NAME setup script on $(date)
export ANTHROPIC_BASE_URL="$API_BASE_URL"
export ANTHROPIC_AUTH_TOKEN="$api_key"
export CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1
# === End $SITE_NAME Configuration ===
EOF

    print_success "$(msg added_shell_env): $shell_config"
}

# ============================================================
# Verify setup / 验证配置
# ============================================================

verify_setup() {
    local api_key="$1"

    print_section "$(msg verify_title)"

    # --- 1. Test API connection ---
    print_info "$(msg testing_api)"

    local test_url="${API_BASE_URL}/v1/models"
    local response_code

    response_code=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "x-api-key: $api_key" \
        -H "anthropic-version: 2023-06-01" \
        --connect-timeout 10 \
        --max-time 30 \
        "$test_url" 2>/dev/null || echo "000")

    if echo "$response_code" | grep -q "^2"; then
        print_success "$(msg api_test_pass) (HTTP $response_code)"
    elif echo "$response_code" | grep -q "^4"; then
        print_warning "$(msg api_test_warn): HTTP $response_code"
        print_info "$(msg api_test_hint)"
    else
        print_warning "$(msg api_test_warn): HTTP $response_code"
        print_info "$(msg api_test_hint)"
    fi

    # --- 2. Verify config file ---
    print_info "$(msg verifying_files)"

    local claude_config="$HOME/.claude/settings.json"
    if [ -f "$claude_config" ]; then
        if grep -q "ANTHROPIC_BASE_URL" "$claude_config"; then
            print_success "$(msg config_verified)"
        else
            print_error "$(msg config_missing)"
            return 1
        fi
    else
        print_error "$(msg config_not_exist): $claude_config"
        return 1
    fi

    return 0
}

# ============================================================
# Show completion message / 显示完成信息
# ============================================================

show_completion() {
    local api_key="$1"

    # Detect shell config file
    local shell_name=$(basename "$SHELL")
    local shell_config=""
    case "$shell_name" in
        bash) shell_config=".bashrc" ;;
        zsh) shell_config=".zshrc" ;;
        *) shell_config=".profile" ;;
    esac

    echo ""
    echo -e "${GREEN}=============================================${NC}"
    echo -e "${GREEN}  $(msg complete_title)${NC}"
    echo -e "${GREEN}=============================================${NC}"
    echo ""

    # Show current configuration
    echo -e "${BOLD}$(msg current_config):${NC}"
    echo -e "  ANTHROPIC_BASE_URL: ${CYAN}$API_BASE_URL${NC}"
    echo -e "  ANTHROPIC_AUTH_TOKEN: ${CYAN}${api_key:0:20}...${NC}"
    echo ""

    # Export to current shell
    export ANTHROPIC_BASE_URL="$API_BASE_URL"
    export ANTHROPIC_AUTH_TOKEN="$api_key"
    export CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1
    print_success "$(msg config_active)"

    echo ""
    echo -e "${BOLD}$(msg next_steps):${NC}"
    echo "  1. $(msg step1): source ~/$shell_config"
    echo "  2. $(msg step2)"
    echo ""
    echo -e "$(msg help_hint): $SERVER_URL"
    echo ""
}

# ============================================================
# Main / 主函数
# ============================================================

main() {
    local api_key=""
    local skip_cleanup=false
    local force=false

    # Parse arguments
    while [ $# -gt 0 ]; do
        case "$1" in
            -k|--key)
                api_key="$2"
                shift 2
                ;;
            --skip-cleanup)
                skip_cleanup=true
                shift
                ;;
            -f|--force)
                force=true
                shift
                ;;
            -h|--help)
                echo "Usage: $0 [-k|--key API_KEY] [--skip-cleanup] [-f|--force]"
                exit 0
                ;;
            *)
                shift
                ;;
        esac
    done

    # 1. Show banner and select language
    show_banner
    select_language
    show_banner

    # 2. Show server info
    print_success "$(msg api_base_url): $API_BASE_URL"

    # 3. Get API Key
    if [ -z "$api_key" ]; then
        echo ""
        print_info "$(msg api_key_hint)"
        print_info "Login: $SERVER_URL -> API Keys -> Create Key"
        echo ""
        read -sp "$(msg api_key_prompt): " api_key < /dev/tty
        echo ""
    fi

    if [ -z "$api_key" ]; then
        print_error "$(msg api_key_empty)"
        exit 1
    fi

    # 4. Environment conflict check
    if ! check_environment_conflicts; then
        if [ "$skip_cleanup" = false ]; then
            echo ""
            if [ "$force" = true ] || confirm "$(msg cleanup_prompt)"; then
                cleanup_conflicts
            else
                print_warning "$(msg skip_cleanup)"
            fi
        fi
    fi

    # 5. Setup Claude Code configuration
    setup_claude_code "$api_key"

    # 6. Verify setup
    if verify_setup "$api_key"; then
        show_completion "$api_key"
    else
        print_error "Configuration verification failed"
        exit 1
    fi
}

main "$@"
