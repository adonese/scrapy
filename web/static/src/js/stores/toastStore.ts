/**
 * Toast Store - Notification queue with auto-dismiss
 *
 * Features:
 * - Multiple toast types (success, error, warning, info)
 * - Auto-dismiss with configurable duration
 * - Queue management
 * - Dismissible notifications
 *
 * Usage:
 * <div x-data>
 *   <button @click="$store.toast.success('Saved successfully!')">
 *     Show Success
 *   </button>
 *
 *   <div class="toast-container">
 *     <template x-for="toast in $store.toast.toasts" :key="toast.id">
 *       <div x-show="toast.visible"
 *            :class="$store.toast.getToastClass(toast.type)">
 *         <span x-text="toast.message"></span>
 *         <button @click="$store.toast.dismiss(toast.id)">✕</button>
 *       </div>
 *     </template>
 *   </div>
 * </div>
 */

type ToastType = 'success' | 'error' | 'warning' | 'info';

interface Toast {
  id: string;
  message: string;
  type: ToastType;
  duration: number;
  visible: boolean;
  timeoutId?: ReturnType<typeof setTimeout>;
}

interface ToastStoreState {
  toasts: Toast[];
  show: (message: string, type: ToastType, duration?: number) => void;
  success: (message: string, duration?: number) => void;
  error: (message: string, duration?: number) => void;
  warning: (message: string, duration?: number) => void;
  info: (message: string, duration?: number) => void;
  dismiss: (id: string) => void;
  clear: () => void;
  getToastClass: (type: ToastType) => string;
}

export function createToastStore() {
  return {
    toasts: [] as Toast[],

    // Show a toast notification
    show(message: string, type: ToastType = 'info', duration: number = 5000) {
      const id = `toast-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;

      const toast: Toast = {
        id,
        message,
        type,
        duration,
        visible: true,
      };

      // Add to queue
      this.toasts.push(toast);

      // Auto-dismiss if duration > 0
      if (duration > 0) {
        toast.timeoutId = setTimeout(() => {
          this.dismiss(id);
        }, duration);
      }

      console.log(`[ToastStore] Showing ${type} toast:`, message);

      // Limit queue size to 5 toasts
      if (this.toasts.length > 5) {
        const removed = this.toasts.shift();
        if (removed?.timeoutId) {
          clearTimeout(removed.timeoutId);
        }
      }
    },

    // Convenience methods for different toast types
    success(message: string, duration: number = 5000) {
      this.show(message, 'success', duration);
    },

    error(message: string, duration: number = 7000) {
      this.show(message, 'error', duration);
    },

    warning(message: string, duration: number = 6000) {
      this.show(message, 'warning', duration);
    },

    info(message: string, duration: number = 5000) {
      this.show(message, 'info', duration);
    },

    // Dismiss a specific toast
    dismiss(id: string) {
      const toast = this.toasts.find((t) => t.id === id);
      if (!toast) return;

      // Clear timeout if exists
      if (toast.timeoutId) {
        clearTimeout(toast.timeoutId);
      }

      // Hide with animation
      toast.visible = false;

      // Remove from queue after animation (300ms)
      setTimeout(() => {
        this.toasts = this.toasts.filter((t) => t.id !== id);
      }, 300);
    },

    // Clear all toasts
    clear() {
      this.toasts.forEach((toast) => {
        if (toast.timeoutId) {
          clearTimeout(toast.timeoutId);
        }
      });
      this.toasts = [];
      console.log('[ToastStore] Cleared all toasts');
    },

    // Get CSS classes for toast type
    getToastClass(type: ToastType): string {
      const baseClasses = 'toast px-4 py-3 rounded-lg shadow-lg flex items-center gap-3 mb-2 transition-all duration-300';

      const typeClasses = {
        success: 'bg-green-500 text-white',
        error: 'bg-red-500 text-white',
        warning: 'bg-amber-500 text-white',
        info: 'bg-blue-500 text-white',
      };

      return `${baseClasses} ${typeClasses[type]}`;
    },

    // Get count of active toasts
    get count(): number {
      return this.toasts.filter((t) => t.visible).length;
    },

    // Check if any toasts are visible
    get hasToasts(): boolean {
      return this.count > 0;
    }
  } as ToastStoreState;
}
