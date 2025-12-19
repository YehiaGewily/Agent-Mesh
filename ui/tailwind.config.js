/** @type {import('tailwindcss').Config} */
export default {
    content: [
        "./index.html",
        "./src/**/*.{js,ts,jsx,tsx}",
    ],
    theme: {
        extend: {
            colors: {
                background: "#0a0a0a",
                magnus: "#00d4ff", // Cyan/Blue
                cedric: "#00ff9d", // Neon Green
                lyra: "#e056fd",   // Purple/Magenta
                card: "#161616",
            },
            fontFamily: {
                sans: ['Inter', 'sans-serif'],
                mono: ['"JetBrains Mono"', 'monospace'],
            },
            boxShadow: {
                'glow-magnus': '0 0 10px rgba(0, 212, 255, 0.3)',
                'glow-cedric': '0 0 10px rgba(0, 255, 157, 0.3)',
                'glow-lyra': '0 0 10px rgba(224, 86, 253, 0.3)',
            }
        },
    },
    plugins: [],
}
