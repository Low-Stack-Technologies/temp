// State management
let selectedFiles = [];
let uploadedFiles = [];
let serverConfig = null;

// Load server configuration
async function loadConfig() {
    try {
        const response = await fetch('/api/config');
        serverConfig = await response.json();
        
        // Update UI with config values
        const uploadLimits = document.getElementById('uploadLimits');
        uploadLimits.textContent = `Upload up to ${serverConfig.max_files_per_upload} files at once • Max ${formatSize(serverConfig.max_upload_size_mb)} per file`;

        // Filter TTL options
        const ttlSelect = document.getElementById('ttlSelect');
        const options = Array.from(ttlSelect.options);
        let hasValidSelection = false;

        options.forEach(option => {
            const value = parseInt(option.value);
            if (value < serverConfig.min_ttl_seconds || value > serverConfig.max_ttl_seconds) {
                option.remove();
            } else {
                if (option.selected) hasValidSelection = true;
            }
        });

        // If current selection is invalid (removed), select the first valid option
        if (!hasValidSelection && ttlSelect.options.length > 0) {
            ttlSelect.selectedIndex = 0;
        }
    } catch (error) {
        console.error('Failed to load config:', error);
        // Fallback to defaults
        serverConfig = {
            max_upload_size_mb: 1024,
            max_files_per_upload: 10,
            min_ttl_seconds: 60,
            max_ttl_seconds: 2592000
        };
    }
}

// Format size in MB/GB
function formatSize(mb) {
    if (mb >= 1024) {
        return `${mb / 1024}GB`;
    }
    return `${mb}MB`;
}

// Initialize on page load
loadConfig();


// DOM elements
const uploadZone = document.getElementById('uploadZone');
const fileInput = document.getElementById('fileInput');
const filesSection = document.getElementById('filesSection');
const filesList = document.getElementById('filesList');
const uploadBtn = document.getElementById('uploadBtn');
const clearBtn = document.getElementById('clearBtn');
const ttlSelect = document.getElementById('ttlSelect');
const uploadedSection = document.getElementById('uploadedSection');
const uploadedList = document.getElementById('uploadedList');

// Event listeners
uploadZone.addEventListener('click', () => fileInput.click());
fileInput.addEventListener('change', handleFileSelect);
uploadBtn.addEventListener('click', uploadFiles);
clearBtn.addEventListener('click', clearFiles);

// Drag and drop
uploadZone.addEventListener('dragover', (e) => {
    e.preventDefault();
    uploadZone.classList.add('drag-over');
});

uploadZone.addEventListener('dragleave', () => {
    uploadZone.classList.remove('drag-over');
});

uploadZone.addEventListener('drop', (e) => {
    e.preventDefault();
    uploadZone.classList.remove('drag-over');
    const files = Array.from(e.dataTransfer.files);
    addFiles(files);
});

// Handle file selection
function handleFileSelect(e) {
    const files = Array.from(e.target.files);
    addFiles(files);
    fileInput.value = ''; // Reset input
}

// Add files to queue
function addFiles(files) {
    if (!serverConfig) {
        alert('Configuration not loaded yet. Please try again.');
        return;
    }
    
    // Validate file count
    const totalFiles = selectedFiles.length + files.length;
    if (totalFiles > serverConfig.max_files_per_upload) {
        alert(`Maximum ${serverConfig.max_files_per_upload} files allowed per upload`);
        return;
    }

    // Validate file sizes
    const maxSize = serverConfig.max_upload_size_mb * 1024 * 1024;
    const invalidFiles = files.filter(f => f.size > maxSize);
    if (invalidFiles.length > 0) {
        alert(`Some files exceed the ${formatSize(serverConfig.max_upload_size_mb)} limit: ${invalidFiles.map(f => f.name).join(', ')}`);
        return;
    }

    // Add files to queue
    files.forEach(file => {
        const fileObj = {
            id: Math.random().toString(36).substr(2, 9),
            file: file,
            progress: 0,
            status: 'pending' // pending, uploading, success, error
        };
        selectedFiles.push(fileObj);
    });

    renderFilesList();
    filesSection.style.display = 'block';
}

// Render files list
function renderFilesList() {
    filesList.innerHTML = '';
    
    selectedFiles.forEach(fileObj => {
        const item = document.createElement('div');
        item.className = 'file-item';
        item.innerHTML = `
            <div class="file-icon">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
                    <path d="M14 2H6C5.46957 2 4.96086 2.21071 4.58579 2.58579C4.21071 2.96086 4 3.46957 4 4V20C4 20.5304 4.21071 21.0391 4.58579 21.4142C4.96086 21.7893 5.46957 22 6 22H18C18.5304 22 19.0391 21.7893 19.4142 21.4142C19.7893 21.0391 20 20.5304 20 20V8L14 2Z" stroke="white" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                    <path d="M14 2V8H20" stroke="white" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
            </div>
            <div class="file-info">
                <div class="file-name">${fileObj.file.name}</div>
                <div class="file-size">${formatFileSize(fileObj.file.size)}</div>
            </div>
            ${renderFileStatus(fileObj)}
            <button class="remove-btn" onclick="removeFile('${fileObj.id}')">
                <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
                    <path d="M15 5L5 15M5 5L15 15" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                </svg>
            </button>
        `;
        filesList.appendChild(item);
    });

    uploadBtn.disabled = selectedFiles.length === 0;
}

// Render file status
function renderFileStatus(fileObj) {
    if (fileObj.status === 'pending') {
        return '';
    }
    
    if (fileObj.status === 'uploading') {
        return `
            <div class="file-progress">
                <div class="progress-bar">
                    <div class="progress-fill" style="width: ${fileObj.progress}%"></div>
                </div>
                <div class="progress-text">${fileObj.progress}%</div>
            </div>
        `;
    }
    
    if (fileObj.status === 'success') {
        return `
            <div class="file-status status-success">
                <svg class="status-icon" viewBox="0 0 20 20" fill="none">
                    <circle cx="10" cy="10" r="9" stroke="currentColor" stroke-width="2"/>
                    <path d="M6 10L9 13L14 7" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                </svg>
                <span>Uploaded</span>
            </div>
        `;
    }
    
    if (fileObj.status === 'error') {
        return `
            <div class="file-status status-error">
                <svg class="status-icon" viewBox="0 0 20 20" fill="none">
                    <circle cx="10" cy="10" r="9" stroke="currentColor" stroke-width="2"/>
                    <path d="M10 6V10M10 14H10.01" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                </svg>
                <span>Failed</span>
            </div>
        `;
    }
}

// Remove file from queue
function removeFile(id) {
    selectedFiles = selectedFiles.filter(f => f.id !== id);
    renderFilesList();
    
    if (selectedFiles.length === 0) {
        filesSection.style.display = 'none';
    }
}

// Clear all files
function clearFiles() {
    selectedFiles = [];
    filesSection.style.display = 'none';
    renderFilesList();
}

// Upload files asynchronously
async function uploadFiles() {
    if (selectedFiles.length === 0) return;

    uploadBtn.disabled = true;
    clearBtn.disabled = true;

    // Get TTL
    const ttlSeconds = parseInt(ttlSelect.value);

    // Upload all files in parallel
    const uploadPromises = selectedFiles.map(fileObj => uploadSingleFile(fileObj, ttlSeconds));
    
    try {
        const results = await Promise.all(uploadPromises);
        
        // Add successful uploads to uploaded list
        results.forEach((result, index) => {
            if (result.success) {
                uploadedFiles.unshift(result.data);
            }
        });
        
        renderUploadedFiles();
        
        // Clear successful uploads from queue
        selectedFiles = selectedFiles.filter(f => f.status !== 'success');
        
        if (selectedFiles.length === 0) {
            filesSection.style.display = 'none';
        } else {
            renderFilesList();
        }
    } catch (error) {
        console.error('Upload error:', error);
    } finally {
        uploadBtn.disabled = false;
        clearBtn.disabled = false;
    }
}

// Upload single file
async function uploadSingleFile(fileObj, ttlSeconds) {
    fileObj.status = 'uploading';
    fileObj.progress = 0;
    renderFilesList();

    const formData = new FormData();
    formData.append('file', fileObj.file);
    formData.append('ttl_seconds', ttlSeconds);

    try {
        const response = await fetch('/api/upload', {
            method: 'POST',
            body: formData
        });

        if (!response.ok) {
            throw new Error(`Upload failed: ${response.statusText}`);
        }

        const data = await response.json();
        
        console.log('Upload response:', data); // Debug log
        
        if (data.success && data.files && data.files.length > 0) {
            fileObj.status = 'success';
            fileObj.progress = 100;
            renderFilesList();
            
            // Construct full download URL from path
            const downloadPath = data.files[0].download_path;
            console.log('Download path:', downloadPath); // Debug log
            const downloadUrl = downloadPath ? `${window.location.origin}${downloadPath}` : undefined;
            console.log('Constructed URL:', downloadUrl); // Debug log
            
            return {
                success: true,
                data: {
                    ...data.files[0],
                    download_url: downloadUrl,
                    expires_at: data.expires_at
                }
            };
        } else {
            throw new Error(data.error || 'Upload failed');
        }
    } catch (error) {
        console.error('Upload error:', error);
        fileObj.status = 'error';
        renderFilesList();
        
        return {
            success: false,
            error: error.message
        };
    }
}

// Render uploaded files
function renderUploadedFiles() {
    if (uploadedFiles.length === 0) {
        uploadedSection.style.display = 'none';
        return;
    }

    uploadedSection.style.display = 'block';
    uploadedList.innerHTML = '';

    uploadedFiles.forEach(file => {
        const item = document.createElement('div');
        item.className = 'uploaded-item';
        
        const expiresAt = new Date(file.expires_at);
        const timeRemaining = getTimeRemaining(expiresAt);
        
        item.innerHTML = `
            <div class="uploaded-header">
                <div class="uploaded-info">
                    <div class="uploaded-name">${file.filename}</div>
                    <div class="uploaded-meta">
                        ${formatFileSize(file.size)} • Expires ${timeRemaining}
                    </div>
                </div>
                <button class="copy-btn" onclick="copyToClipboard('${file.download_url}')">
                    Copy Link
                </button>
            </div>
            <div class="download-url">${file.download_url}</div>
        `;
        
        uploadedList.appendChild(item);
    });
}

// Copy to clipboard
function copyToClipboard(text) {
    navigator.clipboard.writeText(text).then(() => {
        // Show temporary feedback
        const btn = event.target;
        const originalText = btn.textContent;
        btn.textContent = 'Copied!';
        btn.style.background = 'var(--success)';
        btn.style.borderColor = 'var(--success)';
        btn.style.color = 'white';
        
        setTimeout(() => {
            btn.textContent = originalText;
            btn.style.background = '';
            btn.style.borderColor = '';
            btn.style.color = '';
        }, 2000);
    });
}

// Utility functions
function formatFileSize(bytes) {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
}

function getTimeRemaining(date) {
    const now = new Date();
    const diff = date - now;
    
    if (diff < 0) return 'expired';
    
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);
    
    if (days > 0) return `in ${days} day${days > 1 ? 's' : ''}`;
    if (hours > 0) return `in ${hours} hour${hours > 1 ? 's' : ''}`;
    if (minutes > 0) return `in ${minutes} minute${minutes > 1 ? 's' : ''}`;
    return 'soon';
}
