(() => {
  'use strict';
  const button = document.getElementById('download');
  const progress = document.getElementById('progress');
  const status = document.getElementById('status');
  let active = false;
  let readyURL = null;
  let readyFilename = null;
  const mib = n => (n / 1048576).toFixed(1).replace('.', ',');
  const hex = buffer => Array.from(new Uint8Array(buffer), n => n.toString(16).padStart(2, '0')).join('');

  function saveReadyArchive() {
    const link = document.createElement('a');
    link.href = readyURL;
    link.download = readyFilename;
    document.body.appendChild(link);
    link.click();
    link.remove();
  }

  window.addEventListener('unload', () => {
    if (readyURL) URL.revokeObjectURL(readyURL);
  });

  button.addEventListener('click', async () => {
    if (readyURL) { saveReadyArchive(); return; }
    if (active) return;
    active = true;
    button.disabled = true;
    progress.value = 0;
    let url;
    try {
      status.textContent = 'Подготовка архива…';
      const result = await fetch('./manifest.json', {cache:'no-store', credentials:'same-origin'});
      if (!result.ok) throw new Error('Не удалось получить описание архива.');
      const manifest = await result.json();
      if (!Array.isArray(manifest.parts) || !manifest.parts.length) throw new Error('Некорректное описание архива.');
      const blobs = [];
      let loaded = 0;
      for (const part of manifest.parts) {
        const response = await fetch('./' + part.name, {credentials:'same-origin'});
        if (!response.ok) throw new Error('Загрузка прервана. Нажмите кнопку еще раз.');
        const chunks = [];
        let partLoaded = 0;
        if (response.body && response.body.getReader) {
          const reader = response.body.getReader();
          while (true) {
            const {done, value} = await reader.read();
            if (done) break;
            partLoaded += value.byteLength;
            if (partLoaded > part.bytes) { await reader.cancel(); throw new Error('Получен неверный файл архива.'); }
            chunks.push(value);
            loaded += value.byteLength;
            progress.value = Math.min(98, loaded / manifest.bytes * 98);
            status.textContent = 'Получено ' + mib(loaded) + ' из ' + mib(manifest.bytes) + ' МБ…';
          }
        } else {
          const bytes = new Uint8Array(await response.arrayBuffer());
          chunks.push(bytes);
          partLoaded = bytes.byteLength;
          loaded += partLoaded;
        }
        const blob = new Blob(chunks, {type:'application/octet-stream'});
        if (partLoaded !== part.bytes) throw new Error('Файл получен не полностью. Повторите загрузку.');
        if (globalThis.crypto?.subtle) {
          const checksum = hex(await crypto.subtle.digest('SHA-256', await blob.arrayBuffer()));
          if (checksum !== part.sha256) throw new Error('Проверка целостности не пройдена. Повторите загрузку.');
        }
        blobs.push(blob);
      }
      const archive = new Blob(blobs, {type:'application/zip'});
      if (archive.size !== manifest.bytes) throw new Error('Неполный архив.');
      url = URL.createObjectURL(archive);
      readyURL = url;
      readyFilename = manifest.filename;
      saveReadyArchive();
      progress.value = 100;
      button.textContent = 'Сохранить ZIP · 53,2 МБ';
      status.textContent = 'Единый ZIP готов. Если браузер не начал скачивание, нажмите «Сохранить ZIP» еще раз.';
    } catch (error) {
      status.textContent = error.message || 'Не удалось скачать архив. Нажмите кнопку еще раз.';
      if (url) URL.revokeObjectURL(url);
    } finally {
      active = false;
      button.disabled = false;
    }
  });
})();
