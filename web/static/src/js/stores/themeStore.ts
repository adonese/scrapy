/**
 * Theme Store - Dark/light mode with system preference detection
 *
 * Features:
 * - Auto-detects system theme preference
 * - Persists user's theme choice to localStorage
 * - Applies dark mode classes to HTML element
 * - Watches for system theme changes
 *
 * Usage:
 * <button @click="$store.theme.toggle()">
 *   <span x-show="$store.theme.current === 'light'">🌙</span>
 *   <span x-show="$store.theme.current === 'dark'">☀️</span>
 * </button>
 */

import Alpine from 'alpinejs';

type ThemeMode = 'light' | 'dark' | 'system';
type ResolvedTheme = 'light' | 'dark';

interface ThemeStoreState {
  mode: ThemeMode;
  current: ResolvedTheme;
  setTheme: (theme: ThemeMode) => void;
  toggle: () => void;
  init: () => void;
}

export function createThemeStore() {
  return {
    // Persisted theme preference
    mode: Alpine.$persist('system').as('themeMode') as ThemeMode,

    // Currently active theme (resolved)
    current: 'light' as ResolvedTheme,

    // Media query for system preference
    _mediaQuery: null as MediaQueryList | null,

    init() {
      // Set up media query listener
      this._mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');

      // Listen for system theme changes
      this._mediaQuery.addEventListener('change', (e) => {
        if (this.mode === 'system') {
          this._applyTheme(e.matches ? 'dark' : 'light');
        }
      });

      // Apply initial theme
      this._resolveAndApplyTheme();
    },

    // Resolve current theme based on mode
    _resolveAndApplyTheme() {
      let resolved: ResolvedTheme;

      if (this.mode === 'system') {
        resolved = this._mediaQuery?.matches ? 'dark' : 'light';
      } else {
        resolved = this.mode as ResolvedTheme;
      }

      this._applyTheme(resolved);
    },

    // Apply theme to DOM
    _applyTheme(theme: ResolvedTheme) {
      this.current = theme;

      const html = document.documentElement;

      if (theme === 'dark') {
        html.classList.add('dark');
        html.style.colorScheme = 'dark';
      } else {
        html.classList.remove('dark');
        html.style.colorScheme = 'light';
      }

      console.log(`[ThemeStore] Applied theme: ${theme} (mode: ${this.mode})`);
    },

    // Set theme mode
    setTheme(theme: ThemeMode) {
      this.mode = theme;
      this._resolveAndApplyTheme();
    },

    // Toggle between light and dark (ignores system)
    toggle() {
      if (this.mode === 'system') {
        // If on system, toggle to opposite of current
        this.setTheme(this.current === 'dark' ? 'light' : 'dark');
      } else {
        // Toggle between light and dark
        this.setTheme(this.current === 'dark' ? 'light' : 'dark');
      }
    },

    // Check if dark mode is active
    get isDark(): boolean {
      return this.current === 'dark';
    },

    // Check if system mode is active
    get isSystemMode(): boolean {
      return this.mode === 'system';
    },

    // Get system preference
    get systemPreference(): ResolvedTheme {
      return this._mediaQuery?.matches ? 'dark' : 'light';
    }
  } as ThemeStoreState;
}
