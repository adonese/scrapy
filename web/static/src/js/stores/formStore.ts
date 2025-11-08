/**
 * Form Store - Persists form data to localStorage with auto-save
 *
 * Features:
 * - Automatic localStorage persistence using Alpine.persist()
 * - Debounced auto-save (500ms)
 * - Form state management
 * - Reset and clear functionality
 *
 * Usage:
 * <div x-data>
 *   <input x-model="$store.form.data.email" @input="$store.form.save()">
 *   <button @click="$store.form.reset()">Reset</button>
 * </div>
 */

import Alpine from 'alpinejs';

interface FormData {
  [key: string]: any;
}

interface FormStoreState {
  data: FormData;
  isDirty: boolean;
  lastSaved: number | null;
  save: () => void;
  reset: () => void;
  clear: () => void;
  setField: (key: string, value: any) => void;
  getField: (key: string) => any;
}

// Debounce utility
function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: ReturnType<typeof setTimeout> | null = null;

  return function(this: any, ...args: Parameters<T>) {
    const later = () => {
      timeout = null;
      func.apply(this, args);
    };

    if (timeout) {
      clearTimeout(timeout);
    }
    timeout = setTimeout(later, wait);
  };
}

export function createFormStore() {
  return {
    // Persisted form data
    data: Alpine.$persist({}).as('formData'),

    // Form state
    isDirty: false,
    lastSaved: Alpine.$persist(null).as('formLastSaved'),

    // Debounced save function
    _debouncedSave: null as any,

    init() {
      // Initialize debounced save
      this._debouncedSave = debounce(() => {
        this.isDirty = false;
        this.lastSaved = Date.now();
        console.log('[FormStore] Auto-saved form data', this.data);
      }, 500);
    },

    // Save form data (debounced)
    save() {
      this.isDirty = true;
      if (this._debouncedSave) {
        this._debouncedSave();
      }
    },

    // Set a specific field and trigger save
    setField(key: string, value: any) {
      this.data[key] = value;
      this.save();
    },

    // Get a specific field
    getField(key: string) {
      return this.data[key];
    },

    // Reset to last saved state
    reset() {
      this.isDirty = false;
      console.log('[FormStore] Reset to last saved state');
    },

    // Clear all form data
    clear() {
      this.data = {};
      this.isDirty = false;
      this.lastSaved = null;
      console.log('[FormStore] Cleared all form data');
    },

    // Get last saved time as human-readable string
    get lastSavedFormatted(): string {
      if (!this.lastSaved) return 'Never';

      const diff = Date.now() - this.lastSaved;
      const seconds = Math.floor(diff / 1000);
      const minutes = Math.floor(seconds / 60);
      const hours = Math.floor(minutes / 60);

      if (seconds < 60) return 'Just now';
      if (minutes < 60) return `${minutes}m ago`;
      if (hours < 24) return `${hours}h ago`;
      return new Date(this.lastSaved).toLocaleDateString();
    }
  } as FormStoreState;
}
