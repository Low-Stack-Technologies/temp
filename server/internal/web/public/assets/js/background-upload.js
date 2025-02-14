document.getElementById('uploadForm').addEventListener('submit', async (e) => {
  e.preventDefault();
  
  const form = e.target;
  const statusDiv = document.getElementById('uploadStatus');
  const submitButton = form.querySelector('button[type="submit"]');
  
  const formData = new FormData(form);
  
  submitButton.setAttribute('aria-busy', 'true');
  submitButton.disabled = true;
  statusDiv.innerHTML = 'Uploading...';
  
  try {
    const response = await fetch(form.action, {
      method: 'POST',
      body: formData
    });
    
    if (response.ok) {
      const url = await response.text();
      statusDiv.innerHTML = `
        <p>File uploaded successfully!</p>
        <a href="${url}" target="_blank">${url}</a>
      `;
      form.reset();
      document.getElementById('selectedFileName').textContent = '';
    } else {
      const error = await response.json();
      statusDiv.innerHTML = `<p style="color: var(--form-element-invalid-border-color)">Error: ${error.message}</p>`;
    }
  } catch (error) {
    statusDiv.innerHTML = `<p style="color: var(--form-element-invalid-border-color)">Upload failed: ${error.message}</p>`;
  } finally {
    submitButton.removeAttribute('aria-busy');
    submitButton.disabled = false;
  }
});