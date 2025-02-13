const dropZone = document.getElementById('dropZone');
const fileInput = document.getElementById('fileInput');

dropZone.addEventListener('click', () => {
  fileInput.click();
});

dropZone.addEventListener('dragover', (e) => {
  e.preventDefault();
  dropZone.style.borderColor = '#1095c1';
  dropZone.style.backgroundColor = 'rgba(16, 149, 193, 0.1)';
});

dropZone.addEventListener('dragleave', (e) => {
  e.preventDefault();
  dropZone.style.borderColor = '#666';
  dropZone.style.backgroundColor = 'transparent';
});

dropZone.addEventListener('drop', (e) => {
  e.preventDefault();
  fileInput.files = e.dataTransfer.files;
  dropZone.style.borderColor = '#666';
  dropZone.style.backgroundColor = 'transparent';
});

fileInput.addEventListener('change', () => {
  const fileName = fileInput.files[0]?.name || 'No file selected';
  document.getElementById('selectedFileName').textContent = fileName;
});

dropZone.addEventListener('drop', (e) => {
  e.preventDefault();
  fileInput.files = e.dataTransfer.files;
  const fileName = fileInput.files[0]?.name || 'No file selected';
  document.getElementById('selectedFileName').textContent = fileName;
  dropZone.style.borderColor = '#666';
  dropZone.style.backgroundColor = 'transparent';
});