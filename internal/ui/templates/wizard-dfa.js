// wizard-dfa.js - DFA-based wizard controller
//
// This replaces the hardcoded step navigation with DFA state machine logic.
// Instead of managing currentStep internally, all state is managed by the Go DFA backend.

class DFAWizard {
    constructor() {
        this.currentState = null;
        this.stateConfig = null;
        this.wizardData = {};
        this.isInitialized = false;
        this.isBackendAvailable = false;
        
        console.log('DFA Wizard initialized');
    }

    async initialize() {
        console.log('Initializing DFA Wizard...');
        
        // Check if backend is available
        this.isBackendAvailable = !!(window.go && window.go.main && window.go.main.App);
        
        if (this.isBackendAvailable) {
            try {
                // Initialize DFA wizard in backend
                await window.go.main.App.InitializeDFAWizard();
                console.log('DFA Wizard backend initialized');
                
                // Get initial state
                await this.refreshState();
                
                // Setup backend event listeners
                this.setupBackendListeners();
                
            } catch (error) {
                console.error('Failed to initialize DFA wizard backend:', error);
                this.fallbackToLegacyMode();
                return;
            }
        } else {
            console.warn('Backend not available, using simulation mode');
            this.simulateInitialState();
        }
        
        // Setup UI event listeners
        this.setupUIListeners();
        
        this.isInitialized = true;
        console.log('DFA Wizard initialization complete');
    }

    // Get current wizard state from backend
    async refreshState() {
        if (!this.isBackendAvailable) {
            return;
        }
        
        try {
            const stateInfo = await window.go.main.App.GetCurrentWizardState();
            this.currentState = stateInfo.state;
            this.stateConfig = stateInfo.config;
            this.wizardData = stateInfo.data || {};
            
            console.log('State refreshed:', this.currentState, this.stateConfig);
            
            // Update UI to reflect current state
            this.updateUI();
            
        } catch (error) {
            console.error('Failed to refresh wizard state:', error);
        }
    }

    // Update UI based on current DFA state
    updateUI() {
        if (!this.stateConfig) {
            console.warn('No state config available for UI update');
            return;
        }
        
        // Show the appropriate page
        this.showPage(this.currentState);
        
        // Update page content based on state config
        this.updatePageContent();
        
        // Update navigation buttons
        this.updateNavigationButtons();
        
        // Update stepper
        this.updateStepper();
    }

    // Update page content based on state configuration
    updatePageContent() {
        const config = this.stateConfig;
        
        // Update title if available
        if (config.title) {
            const titleElements = document.querySelectorAll('.page-title, .step-title');
            titleElements.forEach(el => el.textContent = config.title);
        }
        
        // Update description if available
        if (config.description) {
            const descElements = document.querySelectorAll('.page-description, .step-description');
            descElements.forEach(el => el.textContent = config.description);
        }
        
        // Handle state-specific content updates
        switch (this.currentState) {
            case 'welcome':
                this.updateWelcomePage();
                break;
            case 'license':
                this.updateLicensePage();
                break;
            case 'components':
                this.updateComponentsPage();
                break;
            case 'location':
                this.updateLocationPage();
                break;
            case 'theme_selection':
                this.updateThemeSelectionPage();
                break;
            case 'ready':
                this.updateReadyPage();
                break;
            case 'installing':
                this.updateInstallingPage();
                break;
            case 'complete':
                this.updateCompletePage();
                break;
        }
    }

    // Update navigation buttons based on available actions
    updateNavigationButtons() {
        const btnNext = document.getElementById('btnNext');
        const btnBack = document.getElementById('btnBack');
        const btnCancel = document.getElementById('btnCancel');
        
        if (!this.stateConfig || !this.stateConfig.actions) {
            return;
        }
        
        // Reset button states
        if (btnNext) {
            btnNext.style.display = 'none';
            btnNext.disabled = true;
            btnNext.textContent = 'Next';
        }
        if (btnBack) {
            btnBack.style.display = 'none';
            btnBack.disabled = true;
        }
        if (btnCancel) {
            btnCancel.style.display = 'none';
            btnCancel.disabled = true;
        }
        
        // Configure buttons based on available actions
        this.stateConfig.actions.forEach(action => {
            switch (action.type) {
                case 'next':
                case 'finish':
                    if (btnNext) {
                        btnNext.style.display = 'block';
                        btnNext.disabled = !action.enabled;
                        btnNext.textContent = action.label || 'Next';
                        btnNext.setAttribute('data-action-id', action.id);
                        if (action.primary) btnNext.classList.add('primary');
                    }
                    break;
                    
                case 'back':
                    if (btnBack) {
                        btnBack.style.display = 'block';
                        btnBack.disabled = !action.enabled;
                        btnBack.textContent = action.label || 'Back';
                        btnBack.setAttribute('data-action-id', action.id);
                    }
                    break;
                    
                case 'cancel':
                    if (btnCancel) {
                        btnCancel.style.display = 'block';
                        btnCancel.disabled = !action.enabled;
                        btnCancel.textContent = action.label || 'Cancel';
                        btnCancel.setAttribute('data-action-id', action.id);
                    }
                    break;
            }
        });
    }

    // Perform a wizard action (next, back, cancel, etc.)
    async performAction(actionType, actionData = {}) {
        if (!this.isInitialized) {
            console.warn('Wizard not initialized, cannot perform action:', actionType);
            return false;
        }
        
        console.log('Performing action:', actionType, 'with data:', actionData);
        
        // Collect form data from current page
        const formData = this.collectCurrentPageData();
        const combinedData = { ...formData, ...actionData };
        
        if (this.isBackendAvailable) {
            try {
                // Send action to backend DFA
                const result = await window.go.main.App.PerformWizardAction(actionType, combinedData);
                
                if (result.success) {
                    // Refresh state after successful action
                    await this.refreshState();
                    return true;
                } else {
                    // Handle validation errors or other issues
                    this.handleActionError(result.error);
                    return false;
                }
                
            } catch (error) {
                console.error('Failed to perform wizard action:', error);
                this.handleActionError(error.toString());
                return false;
            }
        } else {
            // Simulation mode - just fake the navigation
            return this.simulateAction(actionType, combinedData);
        }
    }

    // Collect form data from the currently visible page
    collectCurrentPageData() {
        const data = {};
        const currentPage = document.querySelector('.page.active');
        
        if (!currentPage) return data;
        
        // Collect all form inputs
        const inputs = currentPage.querySelectorAll('input, select, textarea');
        inputs.forEach(input => {
            const name = input.name || input.id;
            if (!name) return;
            
            switch (input.type) {
                case 'checkbox':
                    data[name] = input.checked;
                    break;
                case 'radio':
                    if (input.checked) data[name] = input.value;
                    break;
                default:
                    data[name] = input.value;
            }
        });
        
        // Special handling for component selection
        if (this.currentState === 'components') {
            data.selected_components = this.getSelectedComponents();
        }
        
        console.log('Collected page data:', data);
        return data;
    }

    // Get selected components
    getSelectedComponents() {
        const components = [];
        const checkboxes = document.querySelectorAll('#componentsList input[type="checkbox"]:checked');
        checkboxes.forEach(cb => {
            components.push(cb.value);
        });
        return components;
    }

    // Handle action errors (validation, etc.)
    handleActionError(error) {
        console.error('Wizard action error:', error);
        
        // Show error message to user
        let message = 'An error occurred: ' + error;
        if (error.includes('license')) {
            message = 'Please accept the license agreement to continue.';
        } else if (error.includes('component')) {
            message = 'Please select at least one component to install.';
        } else if (error.includes('path')) {
            message = 'Please specify a valid installation path.';
        }
        
        alert(message);
    }

    // Setup UI event listeners
    setupUIListeners() {
        console.log('Setting up DFA wizard UI listeners');
        
        // Navigation buttons
        const btnNext = document.getElementById('btnNext');
        const btnBack = document.getElementById('btnBack');
        const btnCancel = document.getElementById('btnCancel');
        
        if (btnNext) {
            btnNext.removeEventListener('click', this.handleNext);
            btnNext.addEventListener('click', (e) => this.handleNext(e));
        }
        
        if (btnBack) {
            btnBack.removeEventListener('click', this.handleBack);
            btnBack.addEventListener('click', (e) => this.handleBack(e));
        }
        
        if (btnCancel) {
            btnCancel.removeEventListener('click', this.handleCancel);
            btnCancel.addEventListener('click', (e) => this.handleCancel(e));
        }
        
        // Dynamic form listeners
        this.setupDynamicFormListeners();
    }

    // Setup listeners that need to be refreshed on state changes
    setupDynamicFormListeners() {
        // License checkbox
        const acceptLicense = document.getElementById('acceptLicense');
        if (acceptLicense) {
            acceptLicense.addEventListener('change', (e) => {
                const btnNext = document.getElementById('btnNext');
                if (btnNext) {
                    btnNext.disabled = !e.target.checked;
                }
            });
        }
        
        // Component checkboxes
        const componentCheckboxes = document.querySelectorAll('#componentsList input[type="checkbox"]');
        componentCheckboxes.forEach(cb => {
            cb.addEventListener('change', () => {
                this.updateComponentSelection();
            });
        });
        
        // Theme selection
        const themeSelect = document.getElementById('themeSelect');
        if (themeSelect) {
            themeSelect.addEventListener('change', (e) => {
                this.applyTheme(e.target.value);
            });
        }
        
        // Browse path button
        const browsePath = document.getElementById('browsePath');
        if (browsePath) {
            browsePath.addEventListener('click', () => this.browsePath());
        }
    }

    // Handle Next button click
    async handleNext(e) {
        e.preventDefault();
        const actionId = e.target.getAttribute('data-action-id') || 'next';
        await this.performAction(actionId);
    }

    // Handle Back button click
    async handleBack(e) {
        e.preventDefault();
        const actionId = e.target.getAttribute('data-action-id') || 'back';
        await this.performAction(actionId);
    }

    // Handle Cancel button click
    async handleCancel(e) {
        e.preventDefault();
        
        if (this.currentState === 'installing') {
            if (!confirm('Installation is in progress. Are you sure you want to cancel?')) {
                return;
            }
        } else if (!confirm('Are you sure you want to cancel the installation?')) {
            return;
        }
        
        const actionId = e.target.getAttribute('data-action-id') || 'cancel';
        await this.performAction(actionId);
    }

    // Show a specific page
    showPage(pageName) {
        console.log('DFA Wizard showing page:', pageName);
        
        // Hide all pages
        document.querySelectorAll('.page').forEach(page => {
            page.classList.remove('active');
            page.style.display = 'none';
        });
        
        // Show target page
        const targetPage = document.getElementById('page-' + pageName);
        if (targetPage) {
            targetPage.classList.add('active');
            targetPage.style.display = 'flex';
        } else {
            console.error('Page not found:', pageName);
        }
    }

    // Update stepper component
    updateStepper() {
        // This would be implemented based on your stepper component
        // For now, just log the state change
        console.log('Stepper updated for state:', this.currentState);
        
        const stepsContainer = document.querySelector('.steps');
        if (stepsContainer) {
            stepsContainer.setAttribute('data-current-step', this.currentState);
        }
    }

    // Setup backend event listeners
    setupBackendListeners() {
        if (!window.runtime) return;
        
        // Listen for wizard state changes
        window.runtime.EventsOn('wizard-state-changed', async (data) => {
            console.log('Wizard state changed event:', data);
            await this.refreshState();
        });
        
        // Listen for installation progress
        window.runtime.EventsOn('installation-progress', (data) => {
            this.updateProgress(data);
        });
        
        // Listen for installation completion
        window.runtime.EventsOn('installation-complete', (data) => {
            this.onInstallationComplete(data);
        });
        
        // Listen for validation errors
        window.runtime.EventsOn('wizard-validation-error', (data) => {
            this.handleActionError(data.error);
        });
    }

    // State-specific page updates
    updateWelcomePage() {
        // Update app info if available
        if (this.wizardData.app_name) {
            const titleEl = document.getElementById('appTitle');
            if (titleEl) titleEl.textContent = this.wizardData.app_name + ' Setup';
        }
    }

    updateLicensePage() {
        if (this.wizardData.license_text) {
            const licenseEl = document.getElementById('licenseText');
            if (licenseEl) licenseEl.textContent = this.wizardData.license_text;
        }
    }

    updateComponentsPage() {
        // Components will be updated via backend call to load components
        if (this.isBackendAvailable) {
            this.loadComponents();
        }
    }

    updateLocationPage() {
        if (this.wizardData.install_path) {
            const pathEl = document.getElementById('installPath');
            if (pathEl) pathEl.value = this.wizardData.install_path;
        }
    }

    updateThemeSelectionPage() {
        if (this.wizardData.available_themes) {
            this.loadThemeOptions(this.wizardData.available_themes);
        }
        if (this.wizardData.selected_theme) {
            const themeEl = document.getElementById('themeSelect');
            if (themeEl) themeEl.value = this.wizardData.selected_theme;
        }
    }

    updateReadyPage() {
        // Show installation summary
        const summaryEl = document.getElementById('installSummary');
        if (summaryEl && this.wizardData.install_summary) {
            this.displayInstallSummary(this.wizardData.install_summary);
        }
    }

    updateInstallingPage() {
        // Start installation process
        this.startInstallation();
    }

    updateCompletePage() {
        // Show completion information
        if (this.wizardData.next_steps) {
            this.displayNextSteps(this.wizardData.next_steps);
        }
    }

    // Theme selection specific methods
    loadThemeOptions(themes) {
        const themeSelect = document.getElementById('themeSelect');
        if (!themeSelect || !themes) return;
        
        themeSelect.innerHTML = '';
        themes.forEach(theme => {
            const option = document.createElement('option');
            option.value = theme;
            option.textContent = theme.charAt(0).toUpperCase() + theme.slice(1);
            themeSelect.appendChild(option);
        });
    }

    applyTheme(themeName) {
        document.body.className = '';
        if (themeName && themeName !== 'default') {
            document.body.classList.add('theme-' + themeName);
        }
        
        // Update wizard data
        this.wizardData.selected_theme = themeName;
    }

    // Installation methods
    async startInstallation() {
        if (this.isBackendAvailable) {
            try {
                await window.go.main.App.StartInstallation();
            } catch (error) {
                console.error('Failed to start installation:', error);
            }
        } else {
            // Simulate installation
            this.simulateInstallation();
        }
    }

    updateProgress(data) {
        const progressFill = document.getElementById('progressFill');
        const progressPercent = document.getElementById('progressPercent');
        const progressText = document.getElementById('progressText');
        
        if (progressFill) progressFill.style.width = (data.percent || 0) + '%';
        if (progressPercent) progressPercent.textContent = (data.percent || 0) + '%';
        if (progressText) progressText.textContent = data.status || 'Installing...';
    }

    // Fallback and simulation methods
    fallbackToLegacyMode() {
        console.warn('Falling back to legacy wizard mode');
        // Load the legacy installer.js as fallback
        const script = document.createElement('script');
        script.src = 'installer.js';
        document.head.appendChild(script);
    }

    simulateInitialState() {
        this.currentState = 'welcome';
        this.stateConfig = {
            title: 'Welcome',
            description: 'Welcome to the installation wizard',
            actions: [
                { id: 'next', type: 'next', label: 'Next', enabled: true, primary: true },
                { id: 'cancel', type: 'cancel', label: 'Cancel', enabled: true }
            ]
        };
        this.updateUI();
    }

    async simulateAction(actionType, data) {
        // Simple state progression for simulation
        const stateFlow = {
            'welcome': 'license',
            'license': 'components',
            'components': 'location',
            'location': 'ready',
            'ready': 'installing',
            'installing': 'complete'
        };
        
        if (actionType === 'next' && stateFlow[this.currentState]) {
            this.currentState = stateFlow[this.currentState];
            this.updateUI();
            return true;
        } else if (actionType === 'back') {
            // Simple back navigation
            const backFlow = Object.keys(stateFlow).reduce((acc, key) => {
                acc[stateFlow[key]] = key;
                return acc;
            }, {});
            
            if (backFlow[this.currentState]) {
                this.currentState = backFlow[this.currentState];
                this.updateUI();
                return true;
            }
        }
        
        return false;
    }

    simulateInstallation() {
        let progress = 0;
        const interval = setInterval(() => {
            progress += 10;
            this.updateProgress({
                percent: progress,
                status: `Installing... ${progress}%`
            });
            
            if (progress >= 100) {
                clearInterval(interval);
                setTimeout(() => {
                    this.currentState = 'complete';
                    this.updateUI();
                }, 500);
            }
        }, 500);
    }
}

// Global DFA wizard instance
let dfaWizard = null;

// Initialize DFA wizard when DOM is ready
document.addEventListener('DOMContentLoaded', async () => {
    console.log('Initializing DFA Wizard system...');
    
    dfaWizard = new DFAWizard();
    await dfaWizard.initialize();
    
    // Export for debugging
    window.dfaWizard = dfaWizard;
});

// Export the class for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = DFAWizard;
}