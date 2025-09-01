function toggleNav() {
  var nav = document.getElementById('site-nav');
  nav.classList.toggle('open');
}

// Theme switching functionality
(function() {
  const prefersDarkScheme = window.matchMedia('(prefers-color-scheme: dark)');
  
  // Update logo based on theme
  function updateLogo(theme) {
    const logo = document.querySelector('.brand img');
    if (logo) {
      const currentSrc = logo.getAttribute('src');
      const basePath = currentSrc.replace(/\/(logo|logo-dark)\.svg$/, '');
      const newSrc = theme === 'dark' ? basePath + '/logo-dark.svg' : basePath + '/logo.svg';
      
      if (currentSrc !== newSrc) {
        logo.setAttribute('src', newSrc);
      }
    }
  }
  
  // Initialize theme immediately to prevent flash
  function initTheme() {
    const currentTheme = localStorage.getItem('theme') || 
                        (prefersDarkScheme.matches ? 'dark' : 'light');
    
    document.documentElement.setAttribute('data-theme', currentTheme);
    updateLogo(currentTheme);
    return currentTheme;
  }
  
  // Set initial theme immediately
  const initialTheme = initTheme();
  
  // Set up toggle when DOM is ready
  document.addEventListener('DOMContentLoaded', function() {
    const themeToggle = document.getElementById('theme-toggle');
    
    if (themeToggle) {
      // Set toggle to match current theme
      themeToggle.checked = initialTheme === 'dark';
      
      // Theme toggle event
      themeToggle.addEventListener('change', function() {
        const theme = this.checked ? 'dark' : 'light';
        document.documentElement.setAttribute('data-theme', theme);
        localStorage.setItem('theme', theme);
        updateLogo(theme);
      });
    }
  });
  
  // Listen for system theme changes
  prefersDarkScheme.addEventListener('change', e => {
    if (!localStorage.getItem('theme')) {
      const newTheme = e.matches ? 'dark' : 'light';
      document.documentElement.setAttribute('data-theme', newTheme);
      updateLogo(newTheme);
      const themeToggle = document.getElementById('theme-toggle');
      if (themeToggle) themeToggle.checked = e.matches;
    }
  });
})();
