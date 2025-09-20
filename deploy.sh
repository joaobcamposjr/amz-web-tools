#!/bin/bash

# AMZ Web Tools - Deploy Script
# Usage: ./deploy.sh [start|stop|restart|logs|status]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.prod.yml"
ENV_FILE="env.production"

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

# Check if Docker and Docker Compose are installed
check_dependencies() {
    log_info "Checking dependencies..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    log_success "Dependencies check passed"
}

# Check if environment file exists
check_env_file() {
    if [ ! -f "$ENV_FILE" ]; then
        log_error "Environment file $ENV_FILE not found!"
        log_info "Please create the environment file with your production settings."
        exit 1
    fi
    log_success "Environment file found"
}

# Download Oracle Instant Client
setup_oracle_client() {
    log_info "Setting up Oracle Instant Client..."
    
    # Create directory for Oracle client
    mkdir -p oracle-client
    
    # Check if Oracle client is already downloaded
    if [ -f "oracle-client/oracle-instantclient-basic-linux.x64-21.13.0.0.0dbru.zip" ]; then
        log_info "Oracle Instant Client already downloaded"
        return
    fi
    
    log_warning "Oracle Instant Client not found. You need to download it manually:"
    log_info "1. Go to: https://www.oracle.com/database/technologies/instant-client/linux-x86-64-downloads.html"
    log_info "2. Download: oracle-instantclient-basic-linux.x64-21.13.0.0.0dbru.zip"
    log_info "3. Place it in the oracle-client/ directory"
    log_info "4. Run this script again"
    
    # For now, create a placeholder
    echo "Please download Oracle Instant Client manually" > oracle-client/README.txt
}

# Start services
start_services() {
    log_info "Starting AMZ Web Tools services..."
    
    # Build and start containers
    docker-compose -f $COMPOSE_FILE --env-file $ENV_FILE up -d --build
    
    log_success "Services started successfully!"
    log_info "Frontend: http://localhost"
    log_info "Backend API: http://localhost/api/v1"
    log_info "Health Check: http://localhost/health"
}

# Stop services
stop_services() {
    log_info "Stopping AMZ Web Tools services..."
    
    docker-compose -f $COMPOSE_FILE down
    
    log_success "Services stopped successfully!"
}

# Restart services
restart_services() {
    log_info "Restarting AMZ Web Tools services..."
    
    docker-compose -f $COMPOSE_FILE --env-file $ENV_FILE down
    docker-compose -f $COMPOSE_FILE --env-file $ENV_FILE up -d --build
    
    log_success "Services restarted successfully!"
}

# Show logs
show_logs() {
    log_info "Showing logs..."
    docker-compose -f $COMPOSE_FILE logs -f
}

# Show status
show_status() {
    log_info "Service status:"
    docker-compose -f $COMPOSE_FILE ps
}

# Show help
show_help() {
    echo "AMZ Web Tools - Deploy Script"
    echo ""
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  start     Start all services"
    echo "  stop      Stop all services"
    echo "  restart   Restart all services"
    echo "  logs      Show logs"
    echo "  status    Show service status"
    echo "  help      Show this help message"
    echo ""
    echo "Before first deployment:"
    echo "1. Copy env.production to .env and configure your settings"
    echo "2. Download Oracle Instant Client to oracle-client/ directory"
    echo "3. Run: $0 start"
}

# Main script logic
main() {
    case "${1:-help}" in
        start)
            check_dependencies
            check_env_file
            setup_oracle_client
            start_services
            ;;
        stop)
            stop_services
            ;;
        restart)
            check_dependencies
            check_env_file
            restart_services
            ;;
        logs)
            show_logs
            ;;
        status)
            show_status
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
