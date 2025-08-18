// Go Installer Frontend Application

let currentScreen = 0;
const screens = [
    'welcome-screen',
    'license-screen', 
    'components-screen',
    'path-screen',
    'progress-screen',
    'complete-screen'
];

// Initialize app
window.addEventListener('DOMContentLoaded', () => {
    console.log('Go Installer GUI initialized');
    showScreen(0);
});

// Navigate between screens
function showScreen(index) {
    screens.forEach((screenId, i) => {
        const screen = document.getElementById(screenId);
        if (screen) {
            screen.classList.toggle('active', i === index);
        }
    });
    currentScreen = index;
}

function nextScreen() {
    // Validate current screen before proceeding
    if (currentScreen === 1) { // License screen
        const accepted = document.getElementById('accept-license').checked;
        if (!accepted) {
            alert('Please accept the license agreement to continue.');
            return;
        }
    }
    
    if (currentScreen < screens.length - 1) {
        showScreen(currentScreen + 1);
    }
}

function previousScreen() {
    if (currentScreen > 0) {
        showScreen(currentScreen - 1);
    }
}

// Browse for installation path
function browsePath() {
    // In a real Wails app, this would call a Go function
    if (window.go && window.go.main && window.go.main.App.SelectDirectory) {
        window.go.main.App.SelectDirectory().then(path => {
            if (path) {
                document.getElementById('install-path').value = path;
            }
        });
    } else {
        // Fallback for testing
        const path = prompt('Enter installation path:', document.getElementById('install-path').value);
        if (path) {
            document.getElementById('install-path').value = path;
        }
    }
}

// Start installation
function startInstallation() {
    showScreen(4); // Progress screen
    
    // Simulate installation progress
    simulateInstallation();
    
    // In a real app, this would call Go backend
    if (window.go && window.go.main && window.go.main.App.Install) {
        const config = {
            path: document.getElementById('install-path').value,
            components: getSelectedComponents()
        };
        
        window.go.main.App.Install(config).then(result => {
            if (result.success) {
                showScreen(5); // Complete screen
            } else {
                alert('Installation failed: ' + result.error);
                showScreen(3); // Back to path screen
            }
        });
    }
}

// Get selected components
function getSelectedComponents() {
    const components = [];
    const checkboxes = document.querySelectorAll('#components-screen input[type="checkbox"]:checked');
    checkboxes.forEach(cb => {
        components.push(cb.nextElementSibling.textContent.trim());
    });
    return components;
}

// Simulate installation progress (for demo)
function simulateInstallation() {
    const progressFill = document.getElementById('progress-fill');
    const progressStatus = document.getElementById('progress-status');
    
    const steps = [
        { progress: 0, status: 'Preparing installation...' },
        { progress: 20, status: 'Creating directories...' },
        { progress: 40, status: 'Copying files...' },
        { progress: 60, status: 'Configuring application...' },
        { progress: 80, status: 'Creating shortcuts...' },
        { progress: 100, status: 'Installation complete!' }
    ];
    
    let currentStep = 0;
    
    const interval = setInterval(() => {
        if (currentStep < steps.length) {
            const step = steps[currentStep];
            progressFill.style.width = step.progress + '%';
            progressStatus.textContent = step.status;
            currentStep++;
        } else {
            clearInterval(interval);
            setTimeout(() => {
                showScreen(5); // Complete screen
            }, 500);
        }
    }, 1000);
}

// Finish installation
function finish() {
    const launchApp = document.querySelector('#complete-screen input[type="checkbox"]').checked;
    
    if (window.go && window.go.main && window.go.main.App.Finish) {
        window.go.main.App.Finish(launchApp).then(() => {
            window.close();
        });
    } else {
        // For testing
        alert(launchApp ? 'Launching application...' : 'Installation complete!');
        window.close();
    }
}

// Wails runtime ready event
if (window.runtime) {
    window.runtime.EventsOn('installation-progress', (progress) => {
        const progressFill = document.getElementById('progress-fill');
        const progressStatus = document.getElementById('progress-status');
        
        progressFill.style.width = progress.percent + '%';
        progressStatus.textContent = progress.status;
    });
}
