function secondsToDurationString(seconds) {
  if (seconds === -1) return 'never';
  
  const years = Math.floor(seconds / (365 * 24 * 3600));
  seconds %= (365 * 24 * 3600);
  const days = Math.floor(seconds / (24 * 3600));
  seconds %= (24 * 3600);
  const hours = Math.floor(seconds / 3600);
  seconds %= 3600;
  const minutes = Math.floor(seconds / 60);

  let result = '';
  if (years) result += `${years}y`;
  if (days) result += `${days}d`;
  if (hours) result += `${hours}h`;
  if (minutes) result += `${minutes}m`;
  
  return result || '0m';
}

function generateExpirationOptions() {
  const select = document.getElementById('expirationTime');
  const intervals = [
    { seconds: 1800, label: '30 minutes' },
    { seconds: 3600, label: '1 hour' },
    { seconds: 7200, label: '2 hours' },
    { seconds: 21600, label: '6 hours' },
    { seconds: 43200, label: '12 hours' },
    { seconds: 86400, label: '1 day' },
    { seconds: 172800, label: '2 days' },
    { seconds: 259200, label: '3 days' },
    { seconds: 604800, label: '7 days' },
    { seconds: 1209600, label: '14 days' },
    { seconds: 2592000, label: '30 days' },
    { seconds: 7776000, label: '90 days' },
    { seconds: 31536000, label: '1 year' }
  ];

  // Clear existing options except the first one
  while (select.options.length > 1) {
    select.remove(1);
  }

  // Add options within the allowed range
  intervals.forEach(({ seconds, label }) => {
    if (seconds === -1 && maximumExpirationTimeSeconds === -1) {
      select.add(new Option(label, 'never'));
    } else if (seconds >= minimumExpirationTimeSeconds && 
               seconds <= maximumExpirationTimeSeconds && 
               seconds !== -1) {
      const value = secondsToDurationString(seconds);
      select.add(new Option(label, value));
    }
  });
}

// Generate options when the page loads
generateExpirationOptions();