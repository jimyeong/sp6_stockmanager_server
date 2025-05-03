/**
client code


class ItemService {

  async getItemByBarcode(barcode: string): Promise<Item | null> {
    try {
      const response = await apiClient.get(`/getItem`, {
        params: { barcode }
      });

      if (!response.data || response.status !== 200) {
        return null;
      }

      return response.data as Item;
    } catch (error) {
      console.error('Error fetching item by barcode:', error);
      throw error;
    }
  }

  async stockIn(item: Item, quantity: number, userId: string, notes?: string): Promise<void> {
    try {
      await apiClient.post(`/stockIn`, {
        itemId: item.id,
        quantity,
        userId,
        notes: notes || ''
      });
    } catch (error) {
      console.error('Error performing stock in:', error);
      throw error;
    }
  }

  async stockOut(item: Item, quantity: number, userId: string, notes?: string): Promise<void> {
    try {
      // Validate quantity
      if (quantity > item.quantityInStock) {
        throw new Error('Cannot stock out more than available quantity');
      }

      await apiClient.post(`/stockOut`, {
        itemId: item.id,
        quantity,
        userId,
        notes: notes || ''
      });
    } catch (error) {
      console.error('Error performing stock out:', error);
      throw error;
    }
  }

  // For demo/testing, create a sample item if not found
  async createSampleItemIfNotExists(barcode: string): Promise<Item> {
    try {
      // First try to get the item
      const existingItem = await this.getItemByBarcode(barcode);
      if (existingItem) {
        return existingItem;
      }

      // @@ TODO:

      // If item doesn't exist, create a new one
      const response = await apiClient.post(`/createItem`, {
        barcode,
        name: `Sample Item ${barcode}`,
        description: 'This is a sample item created for testing',
        category: 'Test',
        quantityInStock: 10,
        unitPrice: 9.99
      });

      return response.data as Item;
    } catch (error) {
      console.error('Error creating sample item:', error);

      // Fallback to a local item if API fails (for demo purposes)
      return {
        id: `local-${Date.now()}`,
        barcode,
        name: `Sample Item ${barcode}`,
        description: 'This is a sample item created for testing (local fallback)',
        category: 'Test',
        quantityInStock: 10,
        unitPrice: 9.99,
        lastUpdated: new Date()
      };
    }
  }



  // Register a new item
  async registerItem(item: Item) {
    try {
      const response = await apiClient.post(`/registerItem`, item);
      return response.data;
    } catch (error) {
      console.error('Error registering item:', error);
      throw error;
    }
  }

  // Get all items
  async getItems() {
    try {
      const response = await apiClient.get(`/getItems`);
      return response.data;
    } catch (error) {
      console.error('Error getting items:', error);
      throw error;
    }
  }



}

export default new ItemService();

**/

package apis
