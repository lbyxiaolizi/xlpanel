(() => {
    'use strict';

    const body = document.body;
    const navToggle = document.querySelector('[data-nav-toggle]');
    const sidebarToggle = document.querySelector('[data-sidebar-toggle]');
    const themeToggle = document.querySelector('[data-theme-toggle]');

    if (navToggle) {
        navToggle.addEventListener('click', () => {
            body.classList.toggle('nav-open');
        });
    }

    if (sidebarToggle) {
        sidebarToggle.addEventListener('click', () => {
            body.classList.toggle('sidebar-open');
        });
    }

    document.addEventListener('click', (event) => {
        if (!event.target.closest('.topbar') && body.classList.contains('nav-open')) {
            body.classList.remove('nav-open');
        }
    });

    const applyTheme = (theme) => {
        if (!theme || theme === 'auto') {
            document.documentElement.removeAttribute('data-theme');
            return;
        }
        document.documentElement.setAttribute('data-theme', theme);
    };

    const storedTheme = localStorage.getItem('openhost-theme');
    if (storedTheme) {
        applyTheme(storedTheme);
    }

    if (themeToggle) {
        themeToggle.addEventListener('click', () => {
            const current = document.documentElement.getAttribute('data-theme') || 'light';
            const next = current === 'light' ? 'dark' : 'light';
            applyTheme(next);
            localStorage.setItem('openhost-theme', next);
        });
    }
})();
