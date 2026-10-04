(() => {
  "use strict";

  const body = document.body;
  const progressKey = "go-tutorial-progress-v1";
  const lastPageKey = "go-tutorial-last-page-v1";

  function readProgress() {
    try {
      const value = JSON.parse(localStorage.getItem(progressKey) || "[]");
      return Array.isArray(value) ? value.filter((id) => /^0[1-9]$|^10$/.test(id)) : [];
    } catch (_) {
      return [];
    }
  }

  function saveProgress(progress) {
    try { localStorage.setItem(progressKey, JSON.stringify(progress)); } catch (_) { /* file:// may disable storage */ }
  }

  function paintProgress() {
    const completed = readProgress();
    document.querySelectorAll("[data-progress-label]").forEach((node) => {
      node.textContent = `${completed.length} / 10`;
    });
    document.querySelectorAll("[data-progress-bar]").forEach((node) => {
      node.style.width = `${completed.length * 10}%`;
    });
    document.querySelectorAll("[data-lesson-id]").forEach((node) => {
      const done = completed.includes(node.dataset.lessonId);
      node.classList.toggle("is-done", done);
      if (node.classList.contains("course-link")) {
        node.setAttribute("data-complete", done ? "true" : "false");
      }
    });
    const current = body.dataset.lesson;
    const button = document.querySelector('[data-action="complete"]');
    if (button && current) {
      const done = completed.includes(current);
      button.classList.toggle("is-complete", done);
      const label = button.querySelector("[data-complete-label]");
      if (label) label.textContent = done ? "本章已通关，再点取消" : "我学会了，打个勾";
      button.setAttribute("aria-pressed", done ? "true" : "false");
    }
  }

  paintProgress();

  const completeButton = document.querySelector('[data-action="complete"]');
  if (completeButton) {
    completeButton.addEventListener("click", () => {
      const id = completeButton.dataset.lesson;
      const progress = readProgress();
      const next = progress.includes(id) ? progress.filter((item) => item !== id) : [...progress, id];
      saveProgress(next);
      paintProgress();
    });
  }

  // Remember the last lesson, so local reading progress stays in one browser.
  if (body.dataset.page === "lesson" && body.dataset.lesson && body.dataset.lesson !== "00") {
    try { localStorage.setItem(lastPageKey, body.dataset.lesson); } catch (_) { /* optional */ }
  }

  // Search the left-hand course list.
  document.querySelectorAll("[data-search-lessons]").forEach((input) => {
    input.addEventListener("input", () => {
      const query = input.value.trim().toLocaleLowerCase();
      document.querySelectorAll(".course-link").forEach((link) => {
        const title = `${link.dataset.title || ""} ${link.dataset.lessonId || ""}`.toLocaleLowerCase();
        link.hidden = Boolean(query) && !title.includes(query);
      });
    });
  });

  // Search the learning map and gently fold away stages with no match.
  document.querySelectorAll("[data-search-cards]").forEach((input) => {
    input.addEventListener("input", () => {
      const query = input.value.trim().toLocaleLowerCase();
      let count = 0;
      document.querySelectorAll(".phase-card").forEach((phase) => {
        let phaseCount = 0;
        phase.querySelectorAll(".route-lesson").forEach((link) => {
          const text = `${link.dataset.title || ""} ${link.dataset.lessonId || ""}`.toLocaleLowerCase();
          const matches = !query || text.includes(query);
          link.hidden = !matches;
          if (matches) phaseCount += 1;
        });
        phase.hidden = phaseCount === 0;
        count += phaseCount;
      });
      const empty = document.querySelector("[data-search-empty]");
      if (empty) empty.hidden = count > 0;
    });
  });

  // Flip the little recall hint on every lesson.
  document.querySelectorAll('[data-action="reveal"]').forEach((button) => {
    button.addEventListener("click", () => {
      const answer = button.closest(".memory-note")?.querySelector(".recall-answer");
      if (!answer) return;
      answer.hidden = !answer.hidden;
      button.textContent = answer.hidden ? "翻开提示 ↗" : "收起提示 ↖";
    });
  });

  // Copy any Go / shell / text snippet, including the embedded source files.
  document.querySelectorAll(".copy-code").forEach((button) => {
    button.addEventListener("click", async () => {
      const code = button.closest(".code-card")?.querySelector("pre code")?.textContent || "";
      try {
        if (navigator.clipboard && window.isSecureContext) {
          await navigator.clipboard.writeText(code);
        } else {
          const textarea = document.createElement("textarea");
          textarea.value = code;
          textarea.setAttribute("readonly", "");
          textarea.style.position = "fixed";
          textarea.style.opacity = "0";
          document.body.appendChild(textarea);
          textarea.select();
          document.execCommand("copy");
          textarea.remove();
        }
        button.textContent = "已复制 ✓";
        button.classList.add("is-copied");
      } catch (_) {
        button.textContent = "复制失败";
      }
      window.setTimeout(() => {
        button.textContent = "复制";
        button.classList.remove("is-copied");
      }, 1500);
    });
  });

  // Focus mode keeps the article and hides the two navigation rails.
  document.querySelectorAll('[data-action="focus"]').forEach((button) => {
    button.addEventListener("click", () => {
      body.classList.toggle("focus-mode");
      button.textContent = body.classList.contains("focus-mode") ? "退出专注" : "专注阅读";
    });
  });

  // Mobile drawer for chapter navigation.
  const menuButton = document.querySelector('[data-action="menu"]');
  const scrim = document.querySelector('[data-action="close-menu"]');
  function closeMenu() {
    body.classList.remove("nav-open");
    if (menuButton) menuButton.setAttribute("aria-expanded", "false");
  }
  if (menuButton) {
    menuButton.addEventListener("click", () => {
      const opened = body.classList.toggle("nav-open");
      menuButton.setAttribute("aria-expanded", opened ? "true" : "false");
    });
  }
  if (scrim) scrim.addEventListener("click", closeMenu);
  document.querySelectorAll(".course-link").forEach((link) => link.addEventListener("click", closeMenu));
  document.addEventListener("keydown", (event) => { if (event.key === "Escape") closeMenu(); });

  // Make the table of contents follow the section being read.
  const tocLinks = [...document.querySelectorAll("#toc-nav a")];
  if (tocLinks.length && "IntersectionObserver" in window) {
    const linkById = new Map(tocLinks.map((link) => [decodeURIComponent(link.hash.slice(1)), link]));
    const observer = new IntersectionObserver((entries) => {
      entries.forEach((entry) => {
        if (!entry.isIntersecting) return;
        tocLinks.forEach((link) => link.classList.remove("is-current"));
        linkById.get(entry.target.id)?.classList.add("is-current");
      });
    }, { rootMargin: "-15% 0px -72% 0px" });
    document.querySelectorAll("#article h2[id], #article h3[id]").forEach((heading) => observer.observe(heading));
  }

  // A tiny shuffleable deck for retrieval practice on the home page.
  const flashcards = [
    ["Go 的数组和切片，哪个长度固定？", "数组长度写在类型里、固定不变；切片是对底层数组的一扇可变窗口。"],
    ["`defer` 是先进先出，还是后进先出？", "后进先出（LIFO）：像叠盘子，最后放的先拿。"],
    ["Go 的接口需要显式写 implements 吗？", "不需要。类型只要实现接口要求的方法，就自动满足接口。"],
    ["性能优化的第一步是马上改代码吗？", "先测量：跑 Benchmark 建立基线，再用数据验证优化是否有效。"],
    ["Go 的时间布局里，下午三点写成什么？", "写 `15`。Go 用参考时间 2006-01-02 15:04:05 来表达布局。"],
  ];
  let flashIndex = 0;
  const question = document.querySelector("[data-flash-question]");
  const answer = document.querySelector("[data-flash-answer]");
  const counter = document.querySelector("[data-flash-count]");
  const flip = document.querySelector('[data-action="flip-flash"]');
  const next = document.querySelector('[data-action="next-flash"]');
  function paintFlashcard() {
    if (!question || !answer || !counter) return;
    question.textContent = flashcards[flashIndex][0];
    answer.textContent = flashcards[flashIndex][1];
    answer.hidden = true;
    counter.textContent = `${String(flashIndex + 1).padStart(2, "0")} / ${String(flashcards.length).padStart(2, "0")}`;
    if (flip) flip.textContent = "翻开答案";
  }
  if (flip && answer) {
    flip.addEventListener("click", () => {
      answer.hidden = !answer.hidden;
      flip.textContent = answer.hidden ? "翻开答案" : "收起答案";
    });
  }
  if (next) {
    next.addEventListener("click", () => {
      flashIndex = (flashIndex + 1) % flashcards.length;
      paintFlashcard();
    });
  }
  paintFlashcard();
})();
