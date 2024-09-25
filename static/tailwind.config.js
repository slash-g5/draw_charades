/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./*.{html,js}"],
  safelist: [
    {
      pattern: /border-(red|blue|green|yellow|purple|pink|teal|cyan)-(400)/,
    },
    {
      pattern: /bg-(red|blue|green|yellow|purple|pink|teal|cyan)-(50|100|400)/,
      variants: ['hover'],
    },
    {
      pattern: /text-(red|blue|green|yellow|purple|pink|teal|cyan)-(500)/,
    }
  ],
  theme: {
    extend: {},
  },
  plugins: [],
};
