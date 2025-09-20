#!/bin/bash

# AMZ Web Tools - Server Setup Script
# This script sets up a fresh Ubuntu/CentOS server for deployment

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect OS
detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        VERSION=$VERSION_ID
    elif [ -f /etc/amazon-linux-release ]; then
        OS="amzn"
        VERSION=$(cat /etc/amazon-linux-release | grep -oP '(?<=release )\d+')
    else
        log_error "Cannot detect OS"
        exit 1
    fi
    
    log_info "Detected OS: $OS $VERSION"
}

# Install Docker
install_docker() {
    log_info "Installing Docker..."
    
    case $OS in
        ubuntu|debian)
            # Update package index
            sudo apt-get update
            
            # Install prerequisites
            sudo apt-get install -y \
                apt-transport-https \
                ca-certificates \
                curl \
                gnupg \
                lsb-release
            
            # Add Docker's official GPG key
            curl -fsSL https://download.docker.com/linux/$OS/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
            
            # Set up stable repository
            echo \
                "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/$OS \
                $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
            
            # Install Docker Engine
            sudo apt-get update
            sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
            ;;
            
        centos|rhel|fedora)
            # Install prerequisites
            sudo yum install -y yum-utils
            
            # Add Docker repository
            sudo yum-config-manager \
                --add-repo \
                https://download.docker.com/linux/centos/docker-ce.repo
            
            # Install Docker Engine
            sudo yum install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
            ;;
            
        amzn)
            # Amazon Linux specific installation
            log_info "Installing Docker on Amazon Linux..."
            
            if command -v dnf &> /dev/null; then
                # Amazon Linux 2023 - use Amazon Linux Extras
                sudo dnf update -y
                sudo dnf install -y docker
            else
                # Amazon Linux 2 - use Amazon Linux Extras
                sudo yum update -y
                sudo amazon-linux-extras install -y docker
            fi
            ;;
            
        *)
            log_error "Unsupported OS: $OS"
            exit 1
            ;;
    esac
    
    # Start and enable Docker
    sudo systemctl start docker
    sudo systemctl enable docker
    
    # Add current user to docker group
    sudo usermod -aG docker $USER
    
    log_success "Docker installed successfully"
    log_warning "Please logout and login again for Docker group changes to take effect"
}

# Install Docker Compose (standalone)
install_docker_compose() {
    log_info "Installing Docker Compose..."
    
    # Download latest version
    DOCKER_COMPOSE_VERSION=$(curl -s https://api.github.com/repos/docker/compose/releases/latest | grep 'tag_name' | cut -d\" -f4)
    
    sudo curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    
    sudo chmod +x /usr/local/bin/docker-compose
    
    log_success "Docker Compose installed successfully"
}

# Install Git
install_git() {
    log_info "Installing Git..."
    
    case $OS in
        ubuntu|debian)
            sudo apt-get install -y git
            ;;
        centos|rhel|fedora|amzn)
            if command -v dnf &> /dev/null; then
                sudo dnf install -y git
            else
                sudo yum install -y git
            fi
            ;;
    esac
    
    log_success "Git installed successfully"
}

# Create application directory
setup_app_directory() {
    log_info "Setting up application directory..."
    
    APP_DIR="/opt/amz-web-tools"
    
    if [ ! -d "$APP_DIR" ]; then
        sudo mkdir -p $APP_DIR
        sudo chown $USER:$USER $APP_DIR
    fi
    
    log_success "Application directory created: $APP_DIR"
    log_info "Upload your application files to: $APP_DIR"
}

# Setup firewall
setup_firewall() {
    log_info "Setting up firewall..."
    
    case $OS in
        ubuntu|debian)
            if command -v ufw &> /dev/null; then
                sudo ufw allow 22/tcp    # SSH
                sudo ufw allow 80/tcp    # HTTP
                sudo ufw allow 443/tcp   # HTTPS
                sudo ufw --force enable
            fi
            ;;
        centos|rhel|fedora|amzn)
            if command -v firewall-cmd &> /dev/null; then
                sudo firewall-cmd --permanent --add-service=ssh
                sudo firewall-cmd --permanent --add-service=http
                sudo firewall-cmd --permanent --add-service=https
                sudo firewall-cmd --reload
            fi
            ;;
    esac
    
    log_success "Firewall configured"
}

# Setup SSL certificates (Let's Encrypt)
setup_ssl() {
    log_info "Setting up SSL certificates with Let's Encrypt..."
    
    case $OS in
        ubuntu|debian)
            sudo apt-get install -y certbot
            ;;
        centos|rhel|fedora|amzn)
            if command -v dnf &> /dev/null; then
                sudo dnf install -y certbot
            else
                sudo yum install -y certbot
            fi
            ;;
    esac
    
    log_info "To obtain SSL certificates, run:"
    log_info "sudo certbot certonly --standalone -d yourdomain.com"
    log_info "Then copy certificates to: /opt/amz-web-tools/nginx/ssl/"
}

# Main setup function
main() {
    log_info "Starting AMZ Web Tools server setup..."
    
    detect_os
    install_docker
    install_docker_compose
    install_git
    setup_app_directory
    setup_firewall
    setup_ssl
    
    log_success "Server setup completed successfully!"
    echo ""
    log_info "Next steps:"
    log_info "1. Upload your application files to /opt/amz-web-tools/"
    log_info "2. Configure environment variables in env.production"
    log_info "3. Download Oracle Instant Client"
    log_info "4. Run: ./deploy.sh start"
    echo ""
    log_warning "Don't forget to logout and login again for Docker group changes!"
}

# Run main function
main
