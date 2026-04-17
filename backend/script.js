const fs = require('fs');
const file = '../frontend/src/pages/admin-catalog.astro';
let content = fs.readFileSync(file, 'utf-8');

const newCategoryDropdown = \
              <select name="category" class="dynamic-category-select w-full bg-surface py-2.5 pl-4 pr-10 cursor-pointer rounded-xl text-sm border border-stone-200 focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-all">
                <option value="">Select Category</option>
              </select>\;
              
const oldCategoryDropdown = \
              <select name="category" class="w-full bg-surface py-2.5 pl-4 pr-10 cursor-pointer rounded-xl text-sm border border-stone-200 focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-all">
                <option value="Fiksi Kontemporer">Fiksi Kontemporer</option>
                <option value="Filosofi Modern">Filosofi Modern</option>
                <option value="Teknologi & Sains">Teknologi & Sains</option>
                <option value="Seni & Desain">Seni & Desain</option>
              </select>\.trim();

content = content.split(oldCategoryDropdown).join(newCategoryDropdown);

// Replace Image URL in Add Book
const oldImageField1 = \<div class="space-y-2">
              <label class="text-xs font-bold text-stone-500 uppercase tracking-wider">Image URL</label>
              <input name="image_url" type="url" class="w-full bg-surface py-2.5 px-4 rounded-xl text-sm border border-stone-200 focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-all" placeholder="https://..." />
            </div>\;

const newImageField = \<div class="space-y-2">
              <label class="text-xs font-bold text-stone-500 uppercase tracking-wider">Cover Image</label>
              <input name="image_upload" type="file" accept="image/*" class="w-full bg-surface py-2 px-4 rounded-xl text-sm border border-stone-200 focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-all" />
            </div>\;

content = content.replace(oldImageField1, newImageField);

// Replace Image URL in Edit Book
const oldImageField2 = \<div class="space-y-2">
              <label class="text-xs font-bold text-stone-500 uppercase tracking-wider">Image URL</label>
              <input name="image_url" type="url" class="w-full bg-surface py-2.5 px-4 rounded-xl text-sm border border-stone-200 focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-all" />
            </div>\;
content = content.replace(oldImageField2, newImageField);

const extraFields = \<div class="space-y-2">
              <label class="text-xs font-bold text-stone-500 uppercase tracking-wider">Publisher</label>
              <input name="publisher" type="text" class="w-full bg-surface py-2.5 px-4 rounded-xl text-sm border border-stone-200 focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-all" />
            </div>
            <div class="space-y-2">
              <label class="text-xs font-bold text-stone-500 uppercase tracking-wider">Publication Date</label>
              <input name="publication_date" type="date" class="w-full bg-surface py-2.5 px-4 rounded-xl text-sm border border-stone-200 focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-all" />
            </div>
            <div class="space-y-2">
              <label class="text-xs font-bold text-stone-500 uppercase tracking-wider">Format</label>
              <input name="format" type="text" class="w-full bg-surface py-2.5 px-4 rounded-xl text-sm border border-stone-200 focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-all" />
            </div>
            <div class="space-y-2">
              <label class="text-xs font-bold text-stone-500 uppercase tracking-wider">Pages</label>
              <input name="pages" type="number" min="1" class="w-full bg-surface py-2.5 px-4 rounded-xl text-sm border border-stone-200 focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-all" />
            </div>
            <div class="space-y-2">
              <label class="text-xs font-bold text-stone-500 uppercase tracking-wider">Language</label>
              <input name="language" type="text" class="w-full bg-surface py-2.5 px-4 rounded-xl text-sm border border-stone-200 focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-all" />
            </div>
            <div class="space-y-2">
              <label class="text-xs font-bold text-stone-500 uppercase tracking-wider">Dimensions</label>
              <input name="dimensions" type="text" class="w-full bg-surface py-2.5 px-4 rounded-xl text-sm border border-stone-200 focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-all" />
            </div>
            <div class="space-y-2">
              <label class="text-xs font-bold text-stone-500 uppercase tracking-wider">Weight</label>
              <input name="weight" type="text" class="w-full bg-surface py-2.5 px-4 rounded-xl text-sm border border-stone-200 focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-all" />
            </div>\;

const descBlock = \<div class="md:col-span-2 space-y-2">
              <label class="text-xs font-bold text-stone-500 uppercase tracking-wider">Description</label>\;

content = content.replaceAll(descBlock, extraFields + '\n            ' + descBlock);

// Update JS for categories, form submission (adding uploadImage handling)
content = content.replace(
  "import { getBooks, getBook, createBook, updateBook, deleteBook, formatRupiah, isAuthenticated } from '../lib/api';",
  "import { getBooks, getBook, createBook, updateBook, deleteBook, formatRupiah, isAuthenticated, getCategories, uploadImage } from '../lib/api';"
);

const fetchCatRegex = /async function loadCatalog\(\) \\{[\\s\\S]+?\\}/;
let loadCatBlock = content.match(fetchCatRegex)[0];

const newCatLogic = \
    async function loadCategories() {
      try {
        const categories = await getCategories();
        const selects = document.querySelectorAll('.dynamic-category-select');
        selects.forEach(select => {
          // Keep first option (Select Category)
          const firstOpt = select.options[0].outerHTML;
          const options = categories.map(c => \\\<option value="\">\</option>\\\).join('');
          select.innerHTML = firstOpt + options;
        });
      } catch (err) {
        console.error('Failed to load categories', err);
      }
    }

\;

content = content.replace(loadCatBlock, newCatLogic + loadCatBlock + '\n    loadCategories();\n');

// we also need to rewrite form submission logic, let's inject a new generic script or update manually.
fs.writeFileSync(file, content);
