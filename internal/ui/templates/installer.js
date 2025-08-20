// app.js - Frontend controller for the installer

let currentStep = 'welcome';
let config = null;
let selectedComponents = [];

// Initialize the app when DOM is ready
document.addEventListener('DOMContentLoaded', async () => {
    console.log('Installer UI initialized');
    
    // Set up event listeners FIRST
    setupEventListeners();
    
    // Initialize first page
    showPage('welcome');
    
    // Load configuration from backend (if available)
    if (window.go && window.go.main && window.go.main.App) {
        await loadConfig();
    } else {
        console.log('Running in standalone mode without backend');
        // Set default config for standalone mode
        config = {
            appName: 'My Application',
            version: '1.0.0',
            publisher: 'Example Publisher',
            website: 'https://example.com',
            installPath: 'C:\\Program Files\\MyApp'
        };
    }
});

// Load configuration from backend
async function loadConfig() {
    try {
        config = await window.go.main.App.GetConfig();
        console.log('Config loaded:', config);
        
        // Update UI with config
        document.getElementById('appTitle').textContent = config.appName + ' Setup';
        document.getElementById('appVersion').textContent = 'Version ' + config.version;
        document.getElementById('publisher').textContent = config.publisher || 'Example Publisher';
        document.getElementById('website').textContent = config.website || 'https://example.com';
        // Only update license if config has it
        if (config.license) {
            document.getElementById('licenseText').value = config.license;
        }
        document.getElementById('installPath').value = config.installPath || 'C:\\Program Files\\MyApp';
        
        // Load components
        loadComponents();
        
        // Apply theme if specified
        if (config.theme) {
            document.getElementById('themeSelect').value = config.theme;
            applyTheme(config.theme);
        }
    } catch (error) {
        console.error('Failed to load config:', error);
    }
}

// Load components list
function loadComponents() {
    const container = document.getElementById('componentsList');
    
    // If no config, use the default HTML that's already there
    if (!config || !config.components) {
        // Default components are already in HTML, just setup handlers
        updateComponentSelection();
        return;
    }
    
    // Clear and load from config
    container.innerHTML = '';
    
    config.components.forEach(comp => {
        const item = document.createElement('div');
        item.className = 'component-item';
        
        const sizeText = formatSize(comp.size);
        const isDisabled = comp.required ? 'disabled' : '';
        const isChecked = comp.selected || comp.required ? 'checked' : '';
        
        item.innerHTML = `
            <div class="component-header">
                <div class="component-checkbox">
                    <input type="checkbox" id="comp-${comp.id}" value="${comp.id}" 
                           ${isChecked} ${isDisabled} onchange="updateComponentSelection()">
                    <label for="comp-${comp.id}" class="component-name">${comp.name}</label>
                </div>
                <span class="component-size">${sizeText}</span>
            </div>
            <div class="component-description">${comp.description}</div>
        `;
        
        container.appendChild(item);
        
        if (comp.selected || comp.required) {
            selectedComponents.push(comp.id);
        }
    });
    
    updateSpaceInfo();
}

// Update component selection
function updateComponentSelection() {
    selectedComponents = [];
    const checkboxes = document.querySelectorAll('#componentsList input[type="checkbox"]:checked');
    checkboxes.forEach(cb => {
        selectedComponents.push(cb.value);
    });
    updateSpaceInfo();
}

// Update space information
function updateSpaceInfo() {
    let totalSize = 0;
    config.components.forEach(comp => {
        if (selectedComponents.includes(comp.id) || comp.required) {
            totalSize += comp.size;
        }
    });
    
    document.getElementById('requiredSpace').textContent = formatSize(totalSize);
}

// Format bytes to human readable
function formatSize(bytes) {
    const sizes = ['B', 'KB', 'MB', 'GB'];
    if (bytes === 0) return '0 B';
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
}

// Setup event listeners
function setupEventListeners() {
    console.log('Setting up event listeners');
    
    // Navigation buttons - ENSURE they are bound
    const btnNext = document.getElementById('btnNext');
    const btnBack = document.getElementById('btnBack');
    const btnCancel = document.getElementById('btnCancel');
    
    if (btnNext) {
        btnNext.removeEventListener('click', handleNext); // Remove any existing
        btnNext.addEventListener('click', handleNext);
        console.log('Next button listener attached');
    }
    
    if (btnBack) {
        btnBack.removeEventListener('click', handleBack);
        btnBack.addEventListener('click', handleBack);
        console.log('Back button listener attached');
    }
    
    if (btnCancel) {
        btnCancel.removeEventListener('click', handleCancel);
        btnCancel.addEventListener('click', handleCancel);
        console.log('Cancel button listener attached');
    }
    
    // Theme selector
    const themeSelect = document.getElementById('themeSelect');
    if (themeSelect) {
        themeSelect.addEventListener('change', (e) => {
            applyTheme(e.target.value);
        });
    }
    
    // License checkbox
    const acceptLicense = document.getElementById('acceptLicense');
    if (acceptLicense) {
        acceptLicense.addEventListener('change', (e) => {
            document.getElementById('btnNext').disabled = !e.target.checked;
        });
    }
    
    // Browse button
    const browsePath = document.getElementById('browsePath');
    if (browsePath) {
        browsePath.addEventListener('click', async () => {
            console.log('Browse button clicked');
            
            try {
                if (window.go && window.go.main && window.go.main.App && window.go.main.App.BrowseFolder) {
                    // Use native dialog if available (Wails backend)
                    const path = await window.go.main.App.BrowseFolder();
                    if (path) {
                        document.getElementById('installPath').value = path;
                    }
                } else {
                    // Fallback: Use HTML5 file input
                    console.log('Using HTML5 folder picker fallback');
                    const folderInput = document.getElementById('folderInput');
                    
                    if (folderInput) {
                        // Set up one-time change handler
                        folderInput.onchange = (e) => {
                            const files = e.target.files;
                            if (files && files.length > 0) {
                                // Get the path from the first file and extract directory
                                const fullPath = files[0].webkitRelativePath || files[0].name;
                                const pathParts = fullPath.split('/');
                                
                                if (pathParts.length > 1) {
                                    // Use the folder name from the path
                                    const folderName = pathParts[0];
                                    document.getElementById('installPath').value = `C:\\Program Files\\${folderName}`;
                                } else {
                                    // Fallback to just updating with a generic path
                                    document.getElementById('installPath').value = `C:\\Program Files\\MyApp`;
                                }
                                
                                // Alternative: Let user manually edit the path
                                alert('Folder selected. You can manually edit the installation path if needed.');
                            }
                        };
                        
                        // Trigger the file dialog
                        folderInput.click();
                    } else {
                        // Ultimate fallback: Just let user type the path
                        alert('Please type the installation path manually.\n\nExample: C:\\Program Files\\MyApp');
                        document.getElementById('installPath').focus();
                        document.getElementById('installPath').select();
                    }
                }
            } catch (error) {
                console.error('Failed to browse folder:', error);
                // Fallback to manual input
                alert('Please type the installation path manually.');
                document.getElementById('installPath').focus();
            }
        });
    }
    
    // Listen for backend events (if available)
    if (window.runtime) {
        window.runtime.EventsOn('installation-progress', (data) => {
            updateProgress(data);
        });
        
        window.runtime.EventsOn('installation-complete', (data) => {
            onInstallationComplete(data);
        });
        
        window.runtime.EventsOn('installation-error', (data) => {
            onInstallationError(data);
        });
    }
}

// Apply theme
function applyTheme(themeName) {
    document.body.className = '';
    if (themeName && themeName !== 'default') {
        document.body.classList.add('theme-' + themeName);
    }
}

// Handle next button
async function handleNext() {
    console.log('Next clicked, current step:', currentStep);
    
    switch (currentStep) {
        case 'welcome':
            showPage('license');
            break;
        case 'license':
            if (document.getElementById('acceptLicense').checked) {
                showPage('components');
            } else {
                alert('Please accept the license agreement to continue.');
            }
            break;
        case 'components':
            if (window.go && window.go.main && window.go.main.App) {
                await window.go.main.App.SetSelectedComponents(selectedComponents);
            }
            showPage('path');
            break;
        case 'path':
            const path = document.getElementById('installPath').value;
            if (window.go && window.go.main && window.go.main.App) {
                await window.go.main.App.SetInstallPath(path);
            }
            showPage('install');
            startInstallation();
            break;
        case 'complete':
            await finishInstallation();
            break;
    }
}

// Handle back button
function handleBack() {
    switch (currentStep) {
        case 'license':
            showPage('welcome');
            break;
        case 'components':
            showPage('license');
            break;
        case 'path':
            showPage('components');
            break;
    }
}

// Handle cancel button
async function handleCancel() {
    console.log('Cancel clicked');
    
    if (currentStep === 'install') {
        if (!confirm('Installation is in progress. Are you sure you want to cancel?')) {
            return;
        }
    } else if (currentStep !== 'welcome' && currentStep !== 'complete') {
        if (!confirm('Are you sure you want to cancel the installation?')) {
            return;
        }
    }
    
    try {
        if (window.go && window.go.main && window.go.main.App) {
            await window.go.main.App.ExitInstaller();
        } else {
            // Standalone mode - just close or reload
            if (confirm('Close the installer?')) {
                window.close();
                // If window.close() doesn't work (browser security), reload
                window.location.reload();
            }
        }
    } catch (error) {
        window.close();
    }
}

// Show a specific page
function showPage(pageName) {
    console.log('Showing page:', pageName);
    
    // Hide ALL pages first
    document.querySelectorAll('.page').forEach(page => {
        page.classList.remove('active');
        page.style.display = 'none'; // Force hide
    });
    
    // Show selected page
    const targetPage = document.getElementById('page-' + pageName);
    if (targetPage) {
        targetPage.classList.add('active');
        targetPage.style.display = 'flex'; // Force show
        console.log('Page shown:', pageName);
    } else {
        console.error('Page not found:', pageName);
    }
    
    // Update stepper component
    const stepsContainer = document.querySelector('.steps');
    stepsContainer.setAttribute('data-current-step', pageName);
    
    // Update step indicators
    document.querySelectorAll('.step').forEach(step => {
        step.classList.remove('active');
        step.classList.remove('completed');
    });
    
    // Set active step
    document.querySelector(`.step[data-step="${pageName}"]`).classList.add('active');
    
    // Mark completed steps with animation delay
    const steps = ['welcome', 'license', 'components', 'path', 'install', 'complete'];
    const currentIndex = steps.indexOf(pageName);
    steps.forEach((step, index) => {
        if (index < currentIndex) {
            setTimeout(() => {
                const stepElement = document.querySelector(`.step[data-step="${step}"]`);
                stepElement.classList.add('completed');
            }, index * 100); // Staggered animation
        }
    });
    
    currentStep = pageName;
    updateButtons();
    
    // Add click handlers to completed steps for navigation
    updateStepClickHandlers();
}

// Update navigation buttons
function updateButtons() {
    const btnBack = document.getElementById('btnBack');
    const btnNext = document.getElementById('btnNext');
    const btnCancel = document.getElementById('btnCancel');
    
    switch (currentStep) {
        case 'welcome':
            btnBack.disabled = true;
            btnNext.disabled = false;
            btnNext.textContent = 'Next';
            btnCancel.style.display = 'block';
            break;
        case 'license':
            btnBack.disabled = false;
            btnNext.disabled = !document.getElementById('acceptLicense').checked;
            btnNext.textContent = 'Next';
            break;
        case 'components':
            btnBack.disabled = false;
            btnNext.disabled = false;
            btnNext.textContent = 'Next';
            break;
        case 'path':
            btnBack.disabled = false;
            btnNext.disabled = false;
            btnNext.textContent = 'Install';
            break;
        case 'install':
            btnBack.style.display = 'none';
            btnNext.style.display = 'none';
            btnCancel.textContent = 'Cancel';
            break;
        case 'complete':
            btnBack.style.display = 'none';
            btnNext.disabled = false;
            btnNext.textContent = 'Finish';
            btnNext.style.display = 'block';
            btnCancel.style.display = 'none';
            break;
    }
}

// Start installation
async function startInstallation() {
    const logDiv = document.getElementById('installLog');
    logDiv.innerHTML = 'Starting installation...\n';
    
    if (window.go && window.go.main && window.go.main.App) {
        try {
            await window.go.main.App.StartInstallation({
                path: document.getElementById('installPath').value,
                components: selectedComponents
            });
        } catch (error) {
            console.error('Failed to start installation:', error);
            onInstallationError({ message: error.toString() });
        }
    } else {
        // Simulate installation in standalone mode
        console.log('Simulating installation...');
        let progress = 0;
        const interval = setInterval(() => {
            progress += 10;
            updateProgress({
                percent: progress,
                status: `Installing... ${progress}%`
            });
            
            if (progress >= 100) {
                clearInterval(interval);
                onInstallationComplete({
                    installPath: document.getElementById('installPath').value,
                    components: ['Core', 'Documentation'],
                    duration: '2 minutes'
                });
            }
        }, 500);
    }
}

// Update progress
function updateProgress(data) {
    const progressFill = document.getElementById('progressFill');
    const progressPercent = document.getElementById('progressPercent');
    const progressText = document.getElementById('progressText');
    const logDiv = document.getElementById('installLog');
    
    const percent = Math.round(data.percent || 0);
    progressFill.style.width = percent + '%';
    progressPercent.textContent = percent + '%';
    progressText.textContent = data.status || 'Installing...';
    
    // Add to log
    if (data.status) {
        logDiv.innerHTML += data.status + '\n';
        logDiv.scrollTop = logDiv.scrollHeight;
    }
}

// Handle installation complete
function onInstallationComplete(data) {
    const summaryDiv = document.getElementById('completeSummary');
    
    let summaryHtml = '<ul>';
    if (data.installPath) {
        summaryHtml += `<li>Installation path: ${data.installPath}</li>`;
    }
    if (data.components && data.components.length > 0) {
        summaryHtml += `<li>Installed components: ${data.components.join(', ')}</li>`;
    }
    if (data.duration) {
        summaryHtml += `<li>Installation time: ${data.duration}</li>`;
    }
    summaryHtml += '</ul>';
    
    summaryDiv.innerHTML = summaryHtml;
    
    showPage('complete');
}

// Handle installation error
function onInstallationError(data) {
    alert('Installation failed: ' + (data.message || 'Unknown error'));
    showPage('path'); // Go back to path selection
}

// Finish installation
async function finishInstallation() {
    const launchApp = document.getElementById('launchApp').checked;
    const viewReadme = document.getElementById('viewReadme').checked;
    
    console.log('finishInstallation called');
    console.log('window.go available:', !!window.go);
    console.log('window.runtime available:', !!window.runtime);
    
    try {
        if (window.go && window.go.main && window.go.main.App) {
            console.log('Calling FinishInstallation via Wails');
            await window.go.main.App.FinishInstallation(launchApp, viewReadme);
        } else {
            console.log('Wails backend not available, trying alternatives');
            
            // Try ExitInstaller method
            if (window.go && window.go.main && window.go.main.App && window.go.main.App.ExitInstaller) {
                console.log('Calling ExitInstaller');
                await window.go.main.App.ExitInstaller();
                return;
            }
            
            // Standalone mode fallbacks
            if (launchApp) {
                alert('Application would be launched now.');
            }
            if (viewReadme) {
                alert('README would be opened now.');
            }
            
            // Try multiple exit strategies
            console.log('Trying runtime.Quit()');
            if (window.runtime && window.runtime.Quit) {
                try {
                    window.runtime.Quit();
                    return;
                } catch (e) {
                    console.log('runtime.Quit failed:', e);
                }
            }
            
            console.log('Trying window.close()');
            try {
                window.close();
            } catch (e) {
                console.log('window.close failed:', e);
            }
            
            // Force exit after delay
            setTimeout(() => {
                console.log('Forcing exit with location change');
                try {
                    window.location.href = 'about:blank';
                } catch (e) {
                    console.log('location change failed:', e);
                    alert('Installation completed. Please close this window manually.');
                }
            }, 1000);
        }
    } catch (error) {
        console.error('Error in finishInstallation:', error);
        alert('Installation completed. Please close this window manually.');
        window.close();
    }
}

// Update step click handlers for navigation
function updateStepClickHandlers() {
    const steps = ['welcome', 'license', 'components', 'path', 'install', 'complete'];
    const currentIndex = steps.indexOf(currentStep);
    
    document.querySelectorAll('.step').forEach((stepElement, index) => {
        const stepCircle = stepElement.querySelector('.step-circle');
        
        // Remove existing click handlers
        stepCircle.style.cursor = 'default';
        stepCircle.onclick = null;
        
        // Add click handler only to completed steps (not current or future steps)
        if (index < currentIndex && currentStep !== 'install') {
            stepCircle.style.cursor = 'pointer';
            stepCircle.onclick = () => {
                const targetStep = steps[index];
                navigateToStep(targetStep);
            };
            
            // Add hover tooltip
            stepCircle.title = `Go back to ${steps[index]} step`;
        } else if (index === currentIndex) {
            stepCircle.title = 'Current step';
        } else {
            stepCircle.title = 'Not available yet';
        }
    });
}

// Navigate to a specific step
function navigateToStep(targetStep) {
    // Don't allow navigation during installation
    if (currentStep === 'install') {
        return;
    }
    
    // Check if we can navigate to this step
    const steps = ['welcome', 'license', 'components', 'path', 'install', 'complete'];
    const targetIndex = steps.indexOf(targetStep);
    const currentIndex = steps.indexOf(currentStep);
    
    // Only allow navigation to previous steps
    if (targetIndex < currentIndex) {
        // Special handling for license step
        if (targetStep === 'license' && !document.getElementById('acceptLicense').checked) {
            // Re-check the license if going back to license step
            document.getElementById('acceptLicense').checked = true;
        }
        
        showPage(targetStep);
    }
}

// Add visual feedback for step interaction
function addStepInteractionEffects() {
    document.querySelectorAll('.step').forEach(step => {
        const stepCircle = step.querySelector('.step-circle');
        
        stepCircle.addEventListener('mouseenter', function() {
            if (this.style.cursor === 'pointer') {
                this.style.transform = 'scale(1.1)';
            }
        });
        
        stepCircle.addEventListener('mouseleave', function() {
            if (!step.classList.contains('active')) {
                this.style.transform = 'scale(1)';
            }
        });
    });
}

// Initialize interaction effects when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    addStepInteractionEffects();
});
