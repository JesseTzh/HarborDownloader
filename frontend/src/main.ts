import {createApp} from 'vue'
import App from './App.vue'
import './style.css';

if (import.meta.env.DEV && !(window as any).runtime) {
  ;(window as any).runtime = {
    EventsOnMultiple: () => () => {},
    EventsOff: () => {},
    EventsEmit: () => {},
  }
}

createApp(App).mount('#app')
