/** Russian UI copy. Server-provided domain errors are displayed separately. */
export const ru = {
    appTitle: 'Мини-спортсбук',
    balance: 'Баланс:',
    currency: '₽',
    unavailable: '—',
    outcomes: ['П1', 'Х', 'П2'],
    line: {
        connected: '● Линия обновляется',
        disconnected: '● Связь потеряна, переподключаемся…',
    },
    slip: {
        title: 'Купон',
        priceChanged: 'Коэффициент изменился. Новая цена:',
        confirmPrice: 'Подтвердить коэффициент',
        selectOutcome: 'Выберите исход',
        amount: 'Сумма, ₽',
        sending: 'Отправка…',
        retry: 'Повторить запрос',
        submit: 'Поставить',
    },
    messages: {
        recoveryBlocked:
            'Не удалось восстановить запрос. Новые ставки заблокированы: проверьте хранилище браузера.',
        pendingRecovered:
            'Есть незавершённый запрос. Повторите его для проверки.',
        invalidAmount: 'Введите сумму с точностью до копеек',
        positiveAmount: 'Сумма должна быть положительной',
        persistenceFailed: 'Не удалось сохранить запрос. Ставка не отправлена.',
        serverError: 'Ошибка сервера',
        invalidResponse: 'Некорректный ответ сервера',
        accepted: 'Ставка принята',
        networkError: 'Ошибка сети',
        retryHint: ' Повторите: запрос сохранён, двойного списания не будет.',
        balanceLoadFailed: 'Не удалось загрузить баланс',
    },
} as const;
