import js from '@eslint/js';
import ts from 'typescript-eslint';
import vue from 'eslint-plugin-vue';

export default ts.config(
    { ignores: ['**/node_modules/**', '**/dist/**', '.cache/**'] },
    js.configs.recommended,
    ...ts.configs.recommended,
    ...vue.configs['flat/recommended'],
    {
        files: ['**/*.vue'],
        languageOptions: {
            globals: {
                fetch: 'readonly',
                localStorage: 'readonly',
                EventSource: 'readonly',
                crypto: 'readonly',
            },
            parserOptions: { parser: ts.parser },
        },
        rules: {
            'vue/html-indent': 'off',
            'vue/html-closing-bracket-newline': 'off',
            'vue/html-self-closing': 'off',
            'vue/max-attributes-per-line': 'off',
            'vue/singleline-html-element-content-newline': 'off',
            'vue/multi-word-component-names': 'off',
        },
    },
);
