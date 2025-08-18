// Go Installer Frontend Application

class InstallerApp {
    constructor() {
        this.currentScreen = 0;
        this.screens = [
            'welcome-screen',
            'license-screen',
            'components-screen',
            'path-screen',
            'progress-screen',
            'complete-screen'
        ];
        this.selectedComponents = [];
        this.installPath = 'C:\\Program Files\\GoInstaller';
        
        this.init();
    }

    init() {
        // Load configuration from backend
        this.loadConfig();
        
        // Setup event listeners
        this.setupEventListeners();
        
        // Initialize first screen
        this.showScreen(0);
    }

    setupEventListeners() {
        // License checkbox
        const licenseCheckbox = document.getElementById('accept-license');
        if (licenseCheckbox) {
            licenseCheckbox.addEventListener('change', (e) => {
                const nextBtn = document.getElementById('license-next');
                if (nextBtn) {
                    nextBtn.disabled = !e.target.checked;
                }
            });
        }

        // Install path input
        const pathInput = document.getElementById('install-path');
        if (pathInput) {
            pathInput.addEventListener('change', (e) => {
                this.installPath = e.target.value;
            });
        }
    }

    async loadConfig() {
        // Try to call backend method if Wails runtime is available
        if (window.runtime && window.go && window.go.main && window.go.main.App) {
            try {
                const config = await window.go.main.App.GetInstallConfig();
                this.applyConfig(config);
            } catch (error) {
                console.log('Running without backend:', error);
                this.loadDefaultConfig();
            }
        } else {
            this.loadDefaultConfig();
        }
    }

    loadDefaultConfig() {
        // Default configuration for development
        const config = {
            appName: 'Go Installer',
            version: '1.0.0',
            publisher: 'Go Installer Team',
            license: 'MIT License\n\nPermission is hereby granted...',
            components: [
                { id: 'core', name: 'Core Files', description: 'Required application files', required: true, selected: true, size: 10485760 },
                { id: 'docs', name: 'Documentation', description: 'User manual and API docs', required: false, selected: true, size: 5242880 },
                { id: 'examples', name: 'Examples', description: 'Sample projects', required: false, selected: false, size: 2097152 }
            ],
            installPath: this.installPath
        };
        this.applyConfig(config);
    }

    applyConfig(config) {
        // Apply configuration to UI
        document.querySelector('.version').textContent = `Version ${config.version}`;
        
        // Load license
        const licenseContent = document.getElementById('license-content');
        if (licenseContent) {
            licenseContent.value = config.license;
        }
        
        // Load components
        this.loadComponents(config.components);
        
        // Set default path
        this.installPath = config.installPath;
        const pathInput = document.getElementById('install-path');
        if (pathInput) {
            pathInput.value = this.installPath;
        }
    }

    loadComponents(components) {
        const componentList = document.getElementById('component-list');
        if (!componentList) return;
        
        componentList.innerHTML = '';
        components.forEach(comp => {
            const div = document.createElement('div');
            div.className = 'component';
            div.innerHTML = `
                <label>
                    <input type="checkbox" 
                           data-id="${comp.id}"
                           ${comp.selected ? 'checked' : ''} 
                           ${comp.required ? 'disabled' : ''}>
                    <span>${comp.name} (${this.formatSize(comp.size)})</span>
                </label>
                <div class="component-description">${comp.description}</div>
            `;
            
            const checkbox = div.querySelector('input[type="checkbox"]');
            checkbox.addEventListener('change', (e) => {
                this.updateSelectedComponents();
                this.updateRequiredSpace();
            });
            
            componentList.appendChild(div);
        });
        
        this.updateSelectedComponents();
        this.updateRequiredSpace();
    }

    updateSelectedComponents() {
        this.selectedComponents = [];
        document.querySelectorAll('#component-list input[type="checkbox"]:checked').forEach(cb => {
            this.selectedComponents.push(cb.dataset.id);
        });
    }

    updateRequiredSpace() {
        // This would calculate from actual component sizes
        const requiredSpace = document.getElementById('required-space');
        if (requiredSpace) {
            requiredSpace.textContent = '25 MB'; // Placeholder
        }
        
        // Check available space (would call backend)
        const availableSpace = document.getElementById('available-space');
        if (availableSpace) {
            availableSpace.textContent = '50 GB'; // Placeholder
        }
    }

    formatSize(bytes) {
        const sizes = ['B', 'KB', 'MB', 'GB'];
        if (bytes === 0) return '0 B';
        const i = Math.floor(Math.log(bytes) / Math.log(1024));
        return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
    }

    showScreen(index) {
        // Hide all screens
        this.screens.forEach(screenId => {
            const screen = document.getElementById(screenId);
            if (screen) {
                screen.classList.remove('active');
            }
        });
        
        // Show target screen
        const targetScreen = document.getElementById(this.screens[index]);
        if (targetScreen) {
            targetScreen.classList.add('active');
        }
        
        this.currentScreen = index;
    }

    nextScreen() {
        if (this.currentScreen < this.screens.length - 1) {
            // Validate current screen before proceeding
            if (this.validateCurrentScreen()) {
                this.showScreen(this.currentScreen + 1);
            }
        }
    }

    previousScreen() {
        if (this.currentScreen > 0) {
            this.showScreen(this.currentScreen - 1);
        }
    }

    validateCurrentScreen() {
        switch (this.screens[this.currentScreen]) {
            case 'license-screen':
                const licenseAccepted = document.getElementById('accept-license').checked;
                if (!licenseAccepted) {
                    alert('Please accept the license agreement to continue.');
                    return false;
                }
                break;
            case 'components-screen':
                if (this.selectedComponents.length === 0) {
                    alert('Please select at least one component.');
                    return false;
                }
                break;
            case 'path-screen':
                if (!this.installPath || this.installPath.trim() === '') {
                    alert('Please specify an installation path.');
                    return false;
                }
                break;
        }
        return true;
    }

    async browsePath() {
        // Call backend to open folder dialog
        if (window.runtime && window.go && window.go.main && window.go.main.App) {
            try {
                const path = await window.go.main.App.BrowseFolder();
                if (path) {
                    document.getElementById('install-path').value = path;
                    this.installPath = path;
                }
            } catch (error) {
                console.error('Failed to browse folder:', error);
            }
        } else {
            // Fallback for development
            const path = prompt('Enter installation path:', this.installPath);
            if (path) {
                document.getElementById('install-path').value = path;
                this.installPath = path;
            }
        }
    }

    async startInstallation() {
        // Move to progress screen
        this.showScreen(4); // progress-screen
        
        // Start installation via backend
        if (window.runtime && window.go && window.go.main && window.go.main.App) {
            try {
                const config = {
                    path: this.installPath,
                    components: this.selectedComponents
                };
                
                // Start installation
                await window.go.main.App.StartInstallation(config);
            } catch (error) {
                this.showError('Installation failed: ' + error.message);
            }
        } else {
            // Simulate installation for development
            this.simulateInstallation();
        }
    }

    simulateInstallation() {
        let progress = 0;
        const interval = setInterval(() => {
            progress += 10;
            this.updateProgress(progress, `Installing... ${progress}%`);
            
            if (progress >= 100) {
                clearInterval(interval);
                setTimeout(() => {
                    this.showScreen(5); // complete-screen
                }, 500);
            }
        }, 500);
    }

    updateProgress(percent, status) {
        const progressFill = document.getElementById('progress-fill');
        const progressStatus = document.getElementById('progress-status');
        const progressPercent = document.getElementById('progress-percent');
        
        if (progressFill) {
            progressFill.style.width = percent + '%';
        }
        
        if (progressStatus) {
            progressStatus.textContent = status;
        }
        
        if (progressPercent) {
            progressPercent.textContent = percent + '%';
        }
    }

    showError(message) {
        const errorScreen = document.getElementById('error-screen');
        const errorMessage = document.getElementById('error-message');
        
        if (errorMessage) {
            errorMessage.textContent = message;
        }
        
        // Hide all screens and show error
        this.screens.forEach(screenId => {
            const screen = document.getElementById(screenId);
            if (screen) {
                screen.classList.remove('active');
            }
        });
        
        if (errorScreen) {
            errorScreen.classList.add('active');
        }
    }

    retry() {
        // Go back to path screen
        this.showScreen(3);
    }

    async finish() {
        const launchApp = document.getElementById('launch-app');
        
        if (window.runtime && window.go && window.go.main && window.go.main.App) {
            try {
                await window.go.main.App.FinishInstallation(launchApp ? launchApp.checked : false);
            } catch (error) {
                console.error('Failed to finish installation:', error);
            }
            
            // Close the installer
            if (window.runtime) {
                window.runtime.Quit();
            }
        } else {
            // Development mode
            alert('Installation complete! (Development mode)');
            location.reload();
        }
    }
}

// Initialize app when DOM is ready
let app;
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        app = new InstallerApp();
        window.app = app; // Make it globally accessible
    });
} else {
    app = new InstallerApp();
    window.app = app;
}

// Listen for backend events if available
if (window.runtime) {
    window.runtime.EventsOn('installation-progress', (data) => {
        app.updateProgress(data.percent, data.status);
    });
    
    window.runtime.EventsOn('installation-complete', () => {
        app.showScreen(5); // complete-screen
    });
    
    window.runtime.EventsOn('installation-error', (error) => {
        app.showError(error.message);
    });
}
