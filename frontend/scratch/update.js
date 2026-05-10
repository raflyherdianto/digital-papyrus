const fs = require('fs');

const file = 'src/pages/admin-service.astro';
let content = fs.readFileSync(file, 'utf8');

// The Core Service UI blocks
const coreHTML = `
  <!-- ==================== CORE SERVICE MANAGEMENT ==================== -->
  <div class="flex flex-col md:flex-row items-start md:items-baseline justify-between mb-6 md:mb-8 gap-4 mt-8">
    <div>
      <h2 class="text-2xl md:text-3xl font-extrabold tracking-tight text-primary font-headline mb-2">Service Management</h2>
      <p class="text-stone-500 font-body text-sm max-w-xl">Manage the 'Layanan Kami' offerings displayed on the homepage.</p>
    </div>
    <div class="flex gap-3 md:gap-4 w-full md:w-auto">
      <button id="addCoreBtn" class="flex-1 md:flex-none px-4 md:px-5 py-2.5 bg-primary text-white font-bold rounded-xl text-xs md:text-sm hover:opacity-90 transition-all flex items-center justify-center gap-2 shadow-lg shadow-primary/10">
        <span class="material-symbols-outlined text-sm">add</span> <span class="hidden sm:inline">Add New Service</span>
      </button>
    </div>
  </div>

  <div class="bg-surface-container-lowest rounded-xl p-4 md:p-8 border border-stone-200/10 shadow-sm mb-12">
    <div class="overflow-x-auto">
      <table class="w-full text-left min-w-[700px]">
        <thead>
          <tr class="text-stone-400 font-label text-[10px] uppercase tracking-widest border-b border-stone-200/30">
            <th class="pb-4 md:pb-6 px-3 md:px-4 font-semibold">Service</th>
            <th class="pb-4 md:pb-6 px-3 md:px-4 font-semibold">Description</th>
            <th class="pb-4 md:pb-6 px-3 md:px-4 font-semibold">Status</th>
            <th class="pb-4 md:pb-6 px-3 md:px-4 text-right">Actions</th>
          </tr>
        </thead>
        <tbody id="coreTableBody" class="text-sm font-body">
          <tr id="loadingCoreRow">
            <td colspan="4" class="py-16 text-center">
              <div class="flex flex-col items-center gap-3">
                <span class="material-symbols-outlined text-4xl text-stone-300 animate-spin">progress_activity</span>
                <p class="text-stone-400 text-sm">Loading services...</p>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>

  <!-- Add Core Service Modal -->
  <div id="addCoreModal" class="fixed inset-0 bg-stone-900/50 backdrop-blur-sm z-[9999] hidden flex items-center justify-center opacity-0 transition-opacity duration-300">
    <div class="bg-surface-container-lowest w-full max-w-2xl mx-4 rounded-2xl shadow-xl overflow-hidden transform scale-95 transition-transform duration-300" id="addCoreModalContent">
      <div class="px-6 md:px-8 py-6 border-b border-stone-200/50 flex items-center justify-between">
        <h3 class="text-lg md:text-xl font-bold text-primary font-headline">Add New Service</h3>
        <button class="closeModalBtn text-stone-400 hover:text-stone-600 transition-colors" data-modal="addCoreModal"><span class="material-symbols-outlined">close</span></button>
      </div>
      <div class="px-6 md:px-8 py-6 max-h-[70vh] overflow-y-auto no-scrollbar">
        <div id="addCoreAlert" class="hidden mb-6 p-4 rounded-xl bg-red-50 border border-red-200 text-red-700 text-sm"></div>
        <form id="addCoreForm" class="space-y-6">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4 md:gap-6">
            <div class="md:col-span-2 space-y-2">
              <label class="text-xs font-bold text-stone-500 uppercase tracking-wider">Service Title *</label>
              <input name="title" type="text" required class="w-full bg-surface py-2.5 px-4 rounded-xl text-sm border border-stone-200 focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-all" />
            </div>
            <div class="space-y-2">
              <label class="text-xs font-bold text-stone-500 uppercase tracking-wider">Icon (Material Symbol)</label>
              <input name="icon" type="text" class="w-full bg-surface py-2.5 px-4 rounded-xl text-sm border border-stone-200 focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-all" />
            </div>
            <div class="space-y-2">
              <label class="text-xs font-bold text-stone-500 uppercase tracking-wider">Sort Order</label>
              <input name="sort_order" type="number" min="0" class="w-full bg-surface py-2.5 px-4 rounded-xl text-sm border border-stone-200 focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-all" />
            </div>
            <div class="space-y-2">
              <label class="text-xs font-bold text-stone-500 uppercase tracking-wider">Status</label>
              <select name="is_active" class="w-full bg-surface py-2.5 pl-4 pr-10 cursor-pointer rounded-xl text-sm border border-stone-200 focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-all">
                <option value="true">Active</option>
                <option value="false">Inactive</option>
              </select>
            </div>
            <div class="md:col-span-2 space-y-2">
              <label class="text-xs font-bold text-stone-500 uppercase tracking-wider">Description</label>
              <textarea name="description" class="w-full bg-surface py-3 px-4 rounded-xl text-sm border border-stone-200 focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-all h-20 resize-none"></textarea>
            </div>
          </div>
        </form>
      </div>
      <div class="px-6 md:px-8 py-5 bg-surface-container flex items-center justify-end gap-3 border-t border-stone-200/50">
        <button class="closeModalBtn px-5 py-2.5 text-stone-600 font-bold rounded-xl text-sm hover:bg-stone-200/50 transition-all" data-modal="addCoreModal">Cancel</button>
        <button id="saveCoreBtn" class="px-5 py-2.5 bg-primary text-white font-bold rounded-xl text-sm hover:opacity-90 transition-all shadow-lg shadow-primary/10" type="button">Save Service</button>
      </div>
    </div>
  </div>

  <!-- Edit Core Service Modal -->
  <div id="editCoreModal" class="fixed inset-0 bg-stone-900/50 backdrop-blur-sm z-[9999] hidden flex items-center justify-center opacity-0 transition-opacity duration-300">
    <div class="bg-surface-container-lowest w-full max-w-2xl mx-4 rounded-2xl shadow-xl overflow-hidden transform scale-95 transition-transform duration-300" id="editCoreModalContent">
      <div class="px-6 md:px-8 py-6 border-b border-stone-200/50 flex items-center justify-between">
        <h3 class="text-lg md:text-xl font-bold text-primary font-headline">Edit Service</h3>
        <button class="closeModalBtn text-stone-400 hover:text-stone-600 transition-colors" data-modal="editCoreModal"><span class="material-symbols-outlined">close</span></button>
      </div>
      <div class="px-6 md:px-8 py-6 max-h-[70vh] overflow-y-auto no-scrollbar">
        <div id="editCoreAlert" class="hidden mb-6 p-4 rounded-xl bg-red-50 border border-red-200 text-red-700 text-sm"></div>
        <form id="editCoreForm" class="space-y-6">
          <input type="hidden" name="id" />
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4 md:gap-6">
            <div class="md:col-span-2 space-y-2">
              <label class="text-xs font-bold text-stone-500 uppercase tracking-wider">Service Title *</label>
              <input name="title" type="text" required class="w-full bg-surface py-2.5 px-4 rounded-xl text-sm border border-stone-200 focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-all" />
            </div>
            <div class="space-y-2">
              <label class="text-xs font-bold text-stone-500 uppercase tracking-wider">Icon (Material Symbol)</label>
              <input name="icon" type="text" class="w-full bg-surface py-2.5 px-4 rounded-xl text-sm border border-stone-200 focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-all" />
            </div>
            <div class="space-y-2">
              <label class="text-xs font-bold text-stone-500 uppercase tracking-wider">Sort Order</label>
              <input name="sort_order" type="number" min="0" class="w-full bg-surface py-2.5 px-4 rounded-xl text-sm border border-stone-200 focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-all" />
            </div>
            <div class="space-y-2">
              <label class="text-xs font-bold text-stone-500 uppercase tracking-wider">Status</label>
              <select name="is_active" class="w-full bg-surface py-2.5 pl-4 pr-10 cursor-pointer rounded-xl text-sm border border-stone-200 focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-all">
                <option value="true">Active</option>
                <option value="false">Inactive</option>
              </select>
            </div>
            <div class="md:col-span-2 space-y-2">
              <label class="text-xs font-bold text-stone-500 uppercase tracking-wider">Description</label>
              <textarea name="description" class="w-full bg-surface py-3 px-4 rounded-xl text-sm border border-stone-200 focus:border-primary focus:ring-1 focus:ring-primary outline-none transition-all h-20 resize-none"></textarea>
            </div>
          </div>
        </form>
      </div>
      <div class="px-6 md:px-8 py-5 bg-surface-container flex items-center justify-end gap-3 border-t border-stone-200/50">
        <button class="closeModalBtn px-5 py-2.5 text-stone-600 font-bold rounded-xl text-sm hover:bg-stone-200/50 transition-all" data-modal="editCoreModal">Cancel</button>
        <button id="updateCoreBtn" class="px-5 py-2.5 bg-primary text-white font-bold rounded-xl text-sm hover:opacity-90 transition-all shadow-lg shadow-primary/10" type="button">Save Changes</button>
      </div>
    </div>
  </div>

  <!-- Delete Core Modal -->
  <div id="deleteCoreModal" class="fixed inset-0 bg-stone-900/50 backdrop-blur-sm z-[9999] hidden flex items-center justify-center opacity-0 transition-opacity duration-300">
    <div class="bg-surface-container-lowest w-full max-w-sm mx-4 rounded-2xl shadow-xl overflow-hidden transform scale-95 transition-transform duration-300" id="deleteCoreModalContent">
      <div class="p-6 md:p-8 text-center">
        <div class="w-16 h-16 rounded-full bg-red-50 flex items-center justify-center mx-auto mb-4">
          <span class="material-symbols-outlined text-3xl text-red-500">delete_forever</span>
        </div>
        <h3 class="text-lg font-bold text-primary font-headline mb-2">Delete Service?</h3>
        <p class="text-sm text-stone-500">This action cannot be undone. The service "<span id="deleteCoreTitle" class="font-semibold text-stone-700"></span>" will be permanently removed.</p>
      </div>
      <div class="px-6 md:px-8 py-4 bg-surface-container flex items-center justify-center gap-3 border-t border-stone-200/50">
        <button class="closeModalBtn flex-1 px-5 py-2.5 text-stone-600 font-bold rounded-xl text-sm hover:bg-stone-200/50 transition-all" data-modal="deleteCoreModal">Cancel</button>
        <button id="confirmDeleteCoreBtn" class="flex-1 px-5 py-2.5 bg-red-500 text-white font-bold rounded-xl text-sm hover:bg-red-600 transition-all">Delete</button>
      </div>
    </div>
  </div>

  <!-- ==================== END CORE SERVICE ==================== -->
`;

const jsImportReplacement = `import { getServices, getService, createService, updateService, deleteService, formatRupiah, parseFeatures, isAuthenticated, getCoreServices, createCoreService, updateCoreService, deleteCoreService } from '../lib/api';
    import type { Service, CoreService } from '../lib/api';`;

const jsVariables = `
    let allCoreServices: CoreService[] = [];
    let deleteCoreTargetId = '';
`;

const jsModalsInit = `
    document.querySelectorAll('#viewServiceModal, #addServiceModal, #editServiceModal, #deleteModal, #toast, #addCoreModal, #editCoreModal, #deleteCoreModal').forEach(el => document.body.appendChild(el));
`;

const jsRenderCore = `
    // --- CORE SERVICES LOGIC ---
    async function loadCoreServices() {
      const coreTbody = document.getElementById('coreTableBody')!;
      coreTbody.innerHTML = '<tr><td colspan="4" class="py-16 text-center">Loading...</td></tr>';
      try {
        allCoreServices = await getCoreServices();
        allCoreServices.sort((a,b) => a.sort_order - b.sort_order);
        renderCoreTable();
      } catch(e) {
        coreTbody.innerHTML = '<tr><td colspan="4" class="py-16 text-center text-red-500">Failed to load core services.</td></tr>';
      }
    }

    function renderCoreTable() {
      const coreTbody = document.getElementById('coreTableBody')!;
      if (!allCoreServices.length) {
        coreTbody.innerHTML = '<tr><td colspan="4" class="py-16 text-center">No services found</td></tr>';
        return;
      }

      coreTbody.innerHTML = allCoreServices.map((svc, i) => {
        const border = i < allCoreServices.length - 1 ? 'border-b border-stone-200/10' : '';
        const statusBadge = svc.is_active
          ? '<span class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-[11px] font-bold bg-teal-100 text-teal-800">Active</span>'
          : '<span class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-[11px] font-bold bg-stone-200 text-stone-600">Inactive</span>';

        return \`<tr class="group hover:bg-surface-container-low/50 transition-colors \${border}">
          <td class="py-4 md:py-5 px-3 md:px-4">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center flex-shrink-0">
                <span class="material-symbols-outlined text-primary text-lg">\${svc.icon || 'design_services'}</span>
              </div>
              <p class="font-bold text-primary text-sm">\${svc.title}</p>
            </div>
          </td>
          <td class="py-4 md:py-5 px-3 md:px-4 text-stone-500 max-w-xs truncate" title="\${svc.description}">\${svc.description}</td>
          <td class="py-4 md:py-5 px-3 md:px-4">\${statusBadge}</td>
          <td class="py-4 md:py-5 px-3 md:px-4 text-right">
            <div class="flex items-center justify-end gap-2">
              <button class="editCoreBtn w-8 h-8 rounded-lg flex items-center justify-center text-stone-500 hover:text-primary hover:bg-stone-200/50 transition-all" data-id="\${svc.id}">
                <span class="material-symbols-outlined text-lg">edit</span>
              </button>
              <button class="deleteCoreBtn w-8 h-8 rounded-lg flex items-center justify-center text-stone-500 hover:text-error hover:bg-stone-200/50 transition-all" data-id="\${svc.id}" data-title="\${svc.title}">
                <span class="material-symbols-outlined text-lg">delete</span>
              </button>
            </div>
          </td>
        </tr>\`;
      }).join('');

      coreTbody.querySelectorAll('.editCoreBtn').forEach(btn => {
        btn.addEventListener('click', () => {
          const id = (btn as HTMLElement).dataset.id!;
          const svc = allCoreServices.find(s => s.id === id);
          if (!svc) return;
          const f = document.getElementById('editCoreForm') as HTMLFormElement;
          (f.querySelector('[name="id"]') as HTMLInputElement).value = svc.id;
          (f.querySelector('[name="title"]') as HTMLInputElement).value = svc.title;
          (f.querySelector('[name="icon"]') as HTMLInputElement).value = svc.icon||'';
          (f.querySelector('[name="sort_order"]') as HTMLInputElement).value = String(svc.sort_order||0);
          (f.querySelector('[name="is_active"]') as HTMLSelectElement).value = String(svc.is_active);
          (f.querySelector('[name="description"]') as HTMLTextAreaElement).value = svc.description||'';
          openModal('editCoreModal');
        });
      });

      coreTbody.querySelectorAll('.deleteCoreBtn').forEach(btn => {
        btn.addEventListener('click', () => {
          deleteCoreTargetId = (btn as HTMLElement).dataset.id!;
          document.getElementById('deleteCoreTitle')!.textContent = (btn as HTMLElement).dataset.title||'';
          openModal('deleteCoreModal');
        });
      });
    }

    document.getElementById('addCoreBtn')?.addEventListener('click', () => {
      (document.getElementById('addCoreForm') as HTMLFormElement).reset();
      openModal('addCoreModal');
    });

    document.getElementById('saveCoreBtn')?.addEventListener('click', async () => {
      const f = document.getElementById('addCoreForm') as HTMLFormElement;
      const fd = new FormData(f);
      const title = (fd.get('title') as string||'').trim();
      if (!title) return;
      try {
        await createCoreService({
          title, icon: fd.get('icon') as string, description: fd.get('description') as string,
          sort_order: parseInt(fd.get('sort_order') as string)||0, is_active: fd.get('is_active')!=='false'
        });
        closeModal('addCoreModal'); showToast('Service created!'); await loadCoreServices();
      } catch(e) { alert(e); }
    });

    document.getElementById('updateCoreBtn')?.addEventListener('click', async () => {
      const f = document.getElementById('editCoreForm') as HTMLFormElement;
      const fd = new FormData(f);
      const id = fd.get('id') as string;
      const title = (fd.get('title') as string||'').trim();
      if (!title) return;
      try {
        await updateCoreService(id, {
          title, icon: fd.get('icon') as string, description: fd.get('description') as string,
          sort_order: parseInt(fd.get('sort_order') as string)||0, is_active: fd.get('is_active')!=='false'
        });
        closeModal('editCoreModal'); showToast('Service updated!'); await loadCoreServices();
      } catch(e) { alert(e); }
    });

    document.getElementById('confirmDeleteCoreBtn')?.addEventListener('click', async () => {
      try {
        await deleteCoreService(deleteCoreTargetId);
        closeModal('deleteCoreModal'); showToast('Service deleted!'); await loadCoreServices();
      } catch(e) { alert(e); }
    });
`;

content = content.replace('<AdminLayout title="Package Management | Digital Papyrus" activePage="service">', '<AdminLayout title="Services & Packages | Digital Papyrus" activePage="service">\n' + coreHTML);

content = content.replace(/import \{ getServices, getService, createService, updateService, deleteService, formatRupiah, parseFeatures, isAuthenticated \} from '\.\.\/lib\/api';\s+import type \{ Service \} from '\.\.\/lib\/api';/, jsImportReplacement);

content = content.replace(/let deleteTargetId = '';/, "let deleteTargetId = '';" + jsVariables);

content = content.replace(/document\.querySelectorAll\('#viewServiceModal, #addServiceModal, #editServiceModal, #deleteModal, #toast'\)\.forEach\(el => document\.body\.appendChild\(el\)\);/, jsModalsInit);

content = content.replace(/loadServices\(\);/, 'loadServices();\n    loadCoreServices();\n' + jsRenderCore);

fs.writeFileSync(file, content, 'utf8');
console.log('admin-service.astro updated');
