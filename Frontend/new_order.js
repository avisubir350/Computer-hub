// Global variables
let currentUser = { name: "Service Staff" }; // Mock for quick setup
let serviceLineItems = []; // Array to hold dynamic service objects for the current order

// --- CONSTANTS ---
const DEVICE_TYPES = {
    desktop: "Desktop Computer",
    laptop: "Laptop",
    printer: "Printer",
    toner: "Toner",
    ups: "UPS",
    other: "Others (Specify Below)"
};

const INDIAN_STATES = [
    "Andhra Pradesh", "Arunachal Pradesh", "Assam", "Bihar", "Chhattisgarh",
    "Goa", "Gujarat", "Haryana", "Himachal Pradesh", "Jharkhand",
    "Karnataka", "Kerala", "Madhya Pradesh", "Maharashtra", "Manipur",
    "Meghalaya", "Mizoram", "Nagaland", "Odisha", "Punjab",
    "Rajasthan", "Sikkim", "Tamil Nadu", "Telangana", "Tripura",
    "Uttar Pradesh", "Uttarakhand", "West Bengal" // Targeted Default
];

// MOCK STAFF DATA: Used for the new Engineer dropdown
const MOCK_STAFF = [
    { id: 1, name: "Staff Member A" },
    { id: 2, name: "Staff Member B" },
    { id: 3, name: "Staff Member C" }
];

// --- DOM REFERENCES ---
const form = document.getElementById('new-order-form');
const submitBtn = document.getElementById('submit-order-btn');
const deviceTypeSelect = document.getElementById('deviceType');
const otherDeviceDetailsDiv = document.getElementById('otherDeviceDetails');
const otherDeviceNameInput = document.getElementById('otherDeviceName');
const customerPhoneInput = document.getElementById('customerPhone');
const customerStateSelect = document.getElementById('customerState'); 
const otherAccessoriesCheckbox = document.getElementById('otherAccessoriesCheckbox'); 
const otherAccessoriesDetailsDiv = document.getElementById('otherAccessoriesDetails'); 
const otherAccessoriesNameInput = document.getElementById('otherAccessoriesName');   
const accessoryCheckboxes = () => document.querySelectorAll('.accessory-checkbox:checked'); 
const lineItemsBody = document.getElementById('line-items-body');
const addServiceBtn = document.getElementById('add-service-btn');
const assignedEngineerSelect = document.getElementById('assignedEngineer'); // NEW
const ticketTypeRadios = () => document.querySelectorAll('input[name="ticketType"]:checked'); // NEW
const productStatusRadios = () => document.querySelectorAll('input[name="productStatus"]:checked'); // NEW

// All required fields IDs (UPDATED)
const requiredFields = [
    'customerName', 'customerPhone', 'customerAddress', 
    'customerCity', 'customerState', 'customerZip', 'deviceType', 
    'deviceBrand', 'deviceModelNo', 'issueDescription',
    'assignedEngineer', // NEW REQUIRED
    // 'ticketType' // NEW REQUIRED
];

// --- UTILITY FUNCTIONS ---
function checkAuthentication() {
    const userSession = localStorage.getItem('pcHubUser');
    if (userSession) {
         try { currentUser = JSON.parse(userSession); } catch(e) {}
    }
    document.getElementById('user-display').textContent = currentUser.name || 'Staff';
    return true;
}

function validateIndianPhone(phone) {
    const cleanPhone = phone.replace(/[\s\-\(\)\.]/g, '');
    const indianPhonePatterns = [
        /^(\+91|91)?[6-9]\d{9}$/, 
        /^(\+91|91)?\s?[6-9]\d{4}\s?\d{5}$/, 
    ];
    
    return indianPhonePatterns.some(pattern => pattern.test(cleanPhone)) && 
           (cleanPhone.length >= 10 && cleanPhone.length <= 13);
}

function showNotification(message, type = 'info') {
    const notification = document.createElement('div');
    notification.className = `fixed top-4 right-4 p-4 rounded-lg shadow-lg z-50 ${
        type === 'success' ? 'bg-green-500 text-white' : 'bg-blue-500 text-white'
    }`;
    notification.textContent = message;
    document.body.appendChild(notification);
    setTimeout(() => { notification.remove(); }, 3000);
}

// --- CORE FORM LOGIC ---

/**
 * Calculates and updates all total/discount fields, including the sidebar.
 */
function calculateTotals() {
    let subtotal = 0;
    let totalDiscount = 0;
    
    serviceLineItems.forEach(item => {
        // Ensure rate is treated as a number
        const rate = parseFloat(item.rate) || 0;
        
        // Recalculate item totals based on current values
        // Clamp discount between 0 and 100
        item.discountPercent = Math.max(0, Math.min(100, parseFloat(item.discountPercent) || 0));
        item.discountValue = rate * (item.discountPercent / 100);
        item.finalPrice = rate - item.discountValue;

        subtotal += rate;
        totalDiscount += item.discountValue;
    });

    const grandTotal = subtotal - totalDiscount;

    // Update sidebar summary
    document.getElementById('service-count').textContent = serviceLineItems.length;
    document.getElementById('sidebar-subtotal').textContent = `₹${subtotal.toLocaleString('en-IN')}`;
    document.getElementById('sidebar-discount').textContent = `-₹${Math.round(totalDiscount).toLocaleString('en-IN')}`;
    document.getElementById('sidebar-grand-total').textContent = `₹${Math.round(grandTotal).toLocaleString('en-IN')}`;
    
    validateForm();
}

/**
 * Renders the service line items table.
 */
function renderLineItems() {
    lineItemsBody.innerHTML = ''; // Clear existing rows

    if (serviceLineItems.length === 0) {
        lineItemsBody.innerHTML = `
            <tr>
                <td colspan="5" class="text-center py-4 text-gray-500 italic">No services added yet. Click 'Add Service Line Item'.</td>
            </tr>
        `;
    }

    serviceLineItems.forEach((item, index) => {
        const row = document.createElement('tr');
        row.className = 'hover:bg-gray-50';
        row.innerHTML = `
            <td class="px-3 py-2 whitespace-nowrap text-gray-900">
                <input type="text" data-field="serviceName" data-index="${index}" value="${item.serviceName}" required class="line-item-input w-full p-2 border border-gray-300 rounded-lg text-sm" placeholder="e.g., Motherboard Repair, OS Install">
            </td>
            <td class="px-3 py-2 whitespace-nowrap text-right">
                <input type="number" data-field="rate" data-index="${index}" value="${item.rate}" min="0" required class="line-item-input w-24 p-2 border border-gray-300 rounded-lg text-right">
            </td>
            <td class="px-3 py-2 whitespace-nowrap text-right">
                <input type="number" data-field="discount" data-index="${index}" value="${item.discountPercent}" min="0" max="100" class="line-item-input w-20 p-2 border border-gray-300 rounded-lg text-right">
            </td>
            <td class="px-3 py-2 whitespace-nowrap text-right font-bold text-blue-600">
                ₹${Math.round(item.finalPrice).toLocaleString('en-IN')}
            </td>
            <td class="px-3 py-2 whitespace-nowrap text-right">
                <button type="button" data-index="${index}" class="remove-line-item text-red-600 hover:text-red-900 p-1 rounded-full hover:bg-red-100 transition-colors" title="Remove Service">
                    <i class="ph-trash-fill text-lg"></i>
                </button>
            </td>
        `;
        lineItemsBody.appendChild(row);
    });

    // Add event listeners to the new rows
    setupLineItemEventListeners();
    calculateTotals();
}

function handleLineItemChange(event) {
    const target = event.target;
    const index = parseInt(target.dataset.index);
    let item = serviceLineItems[index];

    if (!item) return;

    const field = target.dataset.field;
    
    if (field === 'serviceName') {
        item.serviceName = target.value.trim();
    } else if (field === 'rate') {
        item.rate = parseFloat(target.value) || 0;
    } else if (field === 'discount') {
        // Clamp between 0 and 100
        item.discountPercent = Math.max(0, Math.min(100, parseFloat(target.value) || 0)); 
        target.value = item.discountPercent; // Update input if clamped
    }

    // Recalculate item totals
    item.discountValue = item.rate * (item.discountPercent / 100);
    item.finalPrice = item.rate - item.discountValue;

    // Update the visual total column without full re-render for performance
    const totalCell = target.closest('tr').querySelector('.text-blue-600');
    if (totalCell) {
        totalCell.textContent = `₹${Math.round(item.finalPrice).toLocaleString('en-IN')}`;
    }

    // CRITICAL: Update the sidebar summary
    calculateTotals();
}

function setupLineItemEventListeners() {
    // Input listeners (Service Name, Rate, Discount)
    lineItemsBody.querySelectorAll('.line-item-input').forEach(input => {
        input.addEventListener('input', handleLineItemChange);
        // Use change for numbers to ensure calculation on blur
        if (input.type === 'number') {
            input.addEventListener('change', handleLineItemChange);
        }
    });

    // Remove button listeners with Confirmation
    lineItemsBody.querySelectorAll('.remove-line-item').forEach(button => {
        button.addEventListener('click', (e) => {
            const index = parseInt(e.currentTarget.dataset.index);
            const itemName = serviceLineItems[index]?.serviceName || 'this service item';

            // CONFIRMATION DIALOG
            if (confirm(`Are you sure you want to remove "${itemName}" from the service list?`)) {
                serviceLineItems.splice(index, 1);
                renderLineItems();
                showNotification(`Service "${itemName}" removed.`, 'info');
            }
        });
    });
}

function addLineItem() {
    const newItem = {
        serviceName: "",
        rate: 500,
        discountPercent: 0,
        discountValue: 0,
        finalPrice: 500,
        isCustom: true,
        notes: ''
    };
    
    serviceLineItems.push(newItem);
    renderLineItems();
}

function handleDeviceTypeChange() {
    if (deviceTypeSelect.value === 'other') {
        otherDeviceDetailsDiv.classList.remove('hidden');
        otherDeviceNameInput.required = true;
    } else {
        otherDeviceDetailsDiv.classList.add('hidden');
        otherDeviceNameInput.required = false;
        otherDeviceNameInput.value = ''; 
    }
    validateForm();
}

function handleOtherAccessoriesChange() {
    if (otherAccessoriesCheckbox.checked) {
        otherAccessoriesDetailsDiv.classList.remove('hidden');
    } else {
        otherAccessoriesDetailsDiv.classList.add('hidden');
        otherAccessoriesNameInput.value = ''; 
    }
}

function handleWarrantyChange() {
    const warrantyDetailsDiv = document.getElementById('warranty-details');
    const underWarranty = document.querySelector('input[name="underWarranty"]:checked')?.value === 'yes';
    
    if (underWarranty) {
        warrantyDetailsDiv.classList.remove('hidden');
        document.getElementById('warrantyNo').required = true;
        document.getElementById('warrantyExpDate').required = true;
    } else {
        warrantyDetailsDiv.classList.add('hidden');
        document.getElementById('warrantyNo').required = false;
        document.getElementById('warrantyExpDate').required = false;
    }
}

/**
 * UPDATED: Includes checks for new required fields (Engineer, Ticket Type).
 */
function validateForm() {
    // 1. Check all standard required fields
    const allFieldsFilled = requiredFields.every(fieldId => {
        const field = document.getElementById(fieldId);
        // Handle select elements, which require a value other than the initial empty option
        if (field && field.tagName === 'SELECT') {
            return field.value.trim() !== '' && field.value.trim() !== '0'; // Check for empty string and default '0' for engineer
        }
        return field && field.value.trim() !== '';
    });
    
    // 1b. Check for the conditional 'otherDeviceName' field
    const otherDetailsValid = (deviceTypeSelect.value !== 'other') || 
                              (otherDeviceNameInput.value.trim() !== '');

    // 2. Validate phone number
    const phoneValid = validateIndianPhone(customerPhoneInput.value);

    // 3. Check for service line items (must have at least one)
    const hasServices = serviceLineItems.length > 0;
    
    // 4. Check Data Backup radio button
    const dataBackupConsent = document.querySelector('input[name="dataBackup"]:checked');
    const dataBackupValid = !!dataBackupConsent;
    
    // 5. Check Ticket Type radio button (required field checked above, but good for redundancy)
    // const ticketTypeValid = ticketTypeRadios().length > 0;
    
    // 6. Check Assigned Engineer (required field checked above, but good for redundancy)
    const engineerAssigned = assignedEngineerSelect.value !== '0';

    // Final button state
    submitBtn.disabled = !(allFieldsFilled && otherDetailsValid && phoneValid && hasServices && dataBackupValid  && engineerAssigned);
}

function setupLogoutButton() {
    const logoutBtn = document.getElementById('logout-btn');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', (e) => {
            e.preventDefault();
            localStorage.removeItem('pcHubUser');
            window.location.href = 'index.html';
        });
    }
}

/**
 * NEW: Populates the engineer dropdown using MOCK_STAFF.
 */
function setupStaffDropdown() {
    const staffSelect = document.getElementById('assignedEngineer');
    if (!staffSelect) return;
    
    // Add default empty option
    staffSelect.innerHTML = '<option value="0" selected disabled>Select Assigned Engineer *</option>';

    // Add staff options
    MOCK_STAFF.forEach(staff => {
        const option = document.createElement('option');
        option.value = staff.id;
        option.textContent = staff.name;
        staffSelect.appendChild(option);
    });

    // Add change listener for validation
    staffSelect.addEventListener('change', validateForm);
}


function setupNewOrderForm() {
    // Inject Device Type Options
    deviceTypeSelect.innerHTML = '<option value="">Select Equipment Type</option>' + 
        Object.entries(DEVICE_TYPES).map(([key, value]) => `<option value="${key}">${value}</option>`).join('');

    // Inject State Options
    customerStateSelect.innerHTML = INDIAN_STATES.map(state => {
        const isWestBengal = state === 'West Bengal' ? 'selected' : '';
        return `<option value="${state}" ${isWestBengal}>${state}</option>`;
    }).join('');
    
    // Inject Staff/Engineer Options (NEW)
    setupStaffDropdown();
    
    // --- ADD EVENT LISTENERS ---
    
    // 1. New Service Line Item Button
    addServiceBtn.addEventListener('click', addLineItem);

    // 2. Required Fields (Input/Change)
    requiredFields.forEach(fieldId => {
        const field = document.getElementById(fieldId);
        if (field && field.tagName !== 'SELECT') {
            field.addEventListener('input', validateForm);
        }
        if (field && (field.tagName === 'SELECT' || field.type === 'date')) {
             field.addEventListener('change', validateForm);
        }
    });

    // 2b. Device Type Change Handler (for showing 'Other' field)
    deviceTypeSelect.addEventListener('change', handleDeviceTypeChange);
    
    // 2c. Other Device Name Input (for validation)
    otherDeviceNameInput.addEventListener('input', validateForm);
    
    // 2d. Other Accessories Checkbox Handler
    otherAccessoriesCheckbox.addEventListener('change', handleOtherAccessoriesChange);
    
    // 3. Data Backup Radios
    document.querySelectorAll('input[name="dataBackup"]').forEach(radio => {
        radio.addEventListener('change', validateForm);
    });
    
    // 4. Ticket Type Radios (NEW)
    document.querySelectorAll('input[name="ticketType"]').forEach(radio => {
        radio.addEventListener('change', validateForm);
    });

    // 5. Warranty Radios
    document.querySelectorAll('input[name="underWarranty"]').forEach(radio => {
        radio.addEventListener('change', handleWarrantyChange);
        radio.addEventListener('change', validateForm);
    });
    
    // 6. Phone number formatting and validation
    customerPhoneInput.addEventListener('input', function() {
        let value = this.value;
        let cleaned = value.replace(/[^\d+]/g, '');
        
        if (cleaned.startsWith('+91')) {
            cleaned = cleaned.substring(3);
            if (cleaned.length <= 10) {
                value = `+91 ${cleaned.slice(0, 5)} ${cleaned.slice(5)}`.trim();
            }
        } else if (cleaned.length <= 10) {
            value = `${cleaned.slice(0, 5)} ${cleaned.slice(5)}`.trim();
        }
        
        this.value = value.trim();
        
        const isValid = validateIndianPhone(this.value);
        if (this.value.length > 0) {
            this.classList.toggle('border-green-300', isValid);
            this.classList.toggle('border-red-300', !isValid);
        } else {
            this.classList.remove('border-red-300', 'border-green-300');
        }
        
        validateForm();
    });

    // 7. Form submission (UPDATED)
    form.addEventListener('submit', (e) => {
        e.preventDefault();
        if (submitBtn.disabled) return;
        
        submitBtn.innerHTML = '<div class="spinner border-t-white mr-2"></div> Creating Order...';
        submitBtn.disabled = true;

        let accessoriesReceived = Array.from(accessoryCheckboxes()).map(cb => cb.value);
        const otherAccessoriesName = otherAccessoriesNameInput.value.trim();
        
        if (otherAccessoriesCheckbox.checked && otherAccessoriesName) {
             accessoriesReceived = accessoriesReceived.filter(item => item !== 'Others (Specify)');
             accessoriesReceived.push(`Other: ${otherAccessoriesName}`);
        } else {
             accessoriesReceived = accessoriesReceived.filter(item => item !== 'Others (Specify)');
        }

        // Calculate final totals
        let subtotal = 0;
        let totalDiscount = 0;
        serviceLineItems.forEach(item => {
            const rate = parseFloat(item.rate) || 0;
            const discountPercent = Math.max(0, Math.min(100, parseFloat(item.discountPercent) || 0));
            const discountValue = rate * (discountPercent / 100);

            subtotal += rate;
            totalDiscount += discountValue;
        });
        const grandTotal = subtotal - totalDiscount;
        
        const deviceTypeKey = document.getElementById('deviceType').value;
        const otherDeviceName = document.getElementById('otherDeviceName').value.trim(); 
        const underWarranty = document.querySelector('input[name="underWarranty"]:checked').value === 'yes';
        const assignedEngineerId = document.getElementById('assignedEngineer').value; // NEW
        const assignedEngineerName = assignedEngineerSelect.options[assignedEngineerSelect.selectedIndex].text.replace(' *', '');
        const productStatusValue = document.querySelector('input[name="productStatus"]:checked')?.value || 'Pending Report'; // NEW

        let finalDeviceType = DEVICE_TYPES[deviceTypeKey];
        if (deviceTypeKey === 'other' && otherDeviceName) {
            finalDeviceType = `Other: ${otherDeviceName}`;
        }
        
        const finalLineItems = serviceLineItems.map(item => ({
            ...item,
            serviceName: item.serviceName.trim() || 'Unspecified Service',
            rate: Math.round(item.rate),
            discountValue: Math.round(item.discountValue),
            finalPrice: Math.round(item.finalPrice)
        }));


        const newTicket = {
            id: 'TICKET-' + Date.now() + '-' + Math.random().toString(36).substr(2, 6).toUpperCase(),
            // Customer Info (for mock storage)
            customerName: document.getElementById('customerName').value.trim(),
            customerPhone: document.getElementById('customerPhone').value.trim(),
            customerAddress: `${document.getElementById('customerAddress').value.trim()}, ${document.getElementById('customerCity').value.trim()}, ${document.getElementById('customerState').value.trim()} - ${document.getElementById('customerZip').value.trim()}`,
            // Device Info
            deviceType: finalDeviceType, 
            deviceBrand: document.getElementById('deviceBrand').value.trim(),
            deviceModelNo: document.getElementById('deviceModelNo').value.trim(),
            deviceSerialNo: document.getElementById('deviceSerialNo').value.trim(),
            issueDescription: document.getElementById('issueDescription').value.trim(),
            devicePassword: document.getElementById('devicePassword').value.trim(), 
            accessoriesReceived: accessoriesReceived,
            
            // ORDER DETAILS (NEW FIELDS)
            ticketType: document.querySelector('input[name="ticketType"]:checked').value, // NEW
            engineerId: assignedEngineerId, // NEW
            engineerName: assignedEngineerName, // NEW
            detailedReport: document.getElementById('detailedReport').value.trim() || null, // NEW
            productStatus: productStatusValue, // NEW

            // Service & Consent
            serviceLineItems: finalLineItems,
            subtotal: Math.round(subtotal),
            totalDiscount: Math.round(totalDiscount),
            totalCost: Math.round(grandTotal),
            dataBackup: document.querySelector('input[name="dataBackup"]:checked').value,
            
            // Warranty & Delivery
            underWarranty: underWarranty,
            warrantyNo: underWarranty ? document.getElementById('warrantyNo').value.trim() : '',
            warrantyExpDate: underWarranty ? document.getElementById('warrantyExpDate').value : '',
            expectedDeliveryDate: document.getElementById('expectedDeliveryDate').value,
            
            // Ticket Status
            status: 'New Order', // Initial Status
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
            createdBy: currentUser.name || 'Staff'
        };

        // Mock Backend: Save to localStorage
        const storedTickets = localStorage.getItem('pcHubTickets');
        const allTickets = storedTickets ? JSON.parse(storedTickets) : [];
        allTickets.unshift(newTicket);
        localStorage.setItem('pcHubTickets', JSON.stringify(allTickets));
        
        // Reset form
        form.reset();
        serviceLineItems = [];
        renderLineItems();
        handleWarrantyChange(); 
        handleDeviceTypeChange(); 
        handleOtherAccessoriesChange(); 
        document.getElementById('customerPhone').classList.remove('border-red-300', 'border-green-300'); 
        
        // Show success and redirect
        setTimeout(() => {
            showNotification('Service order created successfully!', 'success');
            window.location.href = 'dashboard.html';
        }, 1000);
    });

    // Initial updates
    handleWarrantyChange(); 
    handleDeviceTypeChange(); 
    handleOtherAccessoriesChange(); 
    renderLineItems();
}

// Initialize the application when page loads
window.addEventListener('load', () => {
     document.getElementById('loading-screen').classList.add('hidden');
     document.getElementById('main-content').classList.remove('hidden');
     checkAuthentication();
     setupNewOrderForm();
     setupLogoutButton();
});