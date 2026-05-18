import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router';
import { useAuth } from '/src/pages/common/AuthContext';
import { getRecipes } from '../recipes/recipe_apis';
import { getMealPlans } from '../mealplan/mealplan_apis';
import { getGroceryLists } from '../grocery/grocery_apis';
import { FiPlus, FiStar, FiBook, FiCalendar, FiShoppingCart, FiChevronRight } from 'react-icons/fi';

function StatCard({ icon, label, count, to, color }) {
  const Icon = icon;
  return (
    <Link
      to={to}
      className="flex items-center gap-4 bg-white rounded-xl border border-gray-200 shadow-sm p-5 hover:shadow-md hover:border-gray-300 transition group"
    >
      <div className={`p-3 rounded-lg ${color}`}>
        <Icon size={22} className="text-white" />
      </div>
      <div className="flex-1 min-w-0">
        <p className="text-2xl font-bold text-gray-900">{count ?? '—'}</p>
        <p className="text-sm text-gray-500">{label}</p>
      </div>
      <FiChevronRight size={18} className="text-gray-300 group-hover:text-gray-500 transition" />
    </Link>
  );
}

function QuickActions() {
  return (
    <div className="flex flex-wrap gap-3">
      <Link
        to="/recipe/add"
        className="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition"
      >
        <FiPlus size={16} />
        Add Recipe
      </Link>
      <Link
        to="/mealplan/add"
        className="inline-flex items-center gap-2 px-4 py-2 bg-white border border-gray-200 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-50 hover:border-gray-300 transition"
      >
        <FiCalendar size={16} />
        New Meal Plan
      </Link>
      <Link
        to="/grocery/add"
        className="inline-flex items-center gap-2 px-4 py-2 bg-white border border-gray-200 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-50 hover:border-gray-300 transition"
      >
        <FiShoppingCart size={16} />
        New Grocery List
      </Link>
    </div>
  );
}

function FavouriteRecipes({ recipes }) {
  const favourites = recipes?.filter(r => r.is_favourite) ?? [];

  if (!recipes) return null;

  if (favourites.length === 0) {
    return (
      <div className="bg-white rounded-xl border border-gray-200 shadow-sm p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-3 flex items-center gap-2">
          <FiStar size={18} className="text-yellow-400" />
          Favourites
        </h2>
        <p className="text-sm text-gray-400">
          No favourites yet — star a recipe to pin it here.
        </p>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-xl border border-gray-200 shadow-sm p-6">
      <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
        <FiStar size={18} className="text-yellow-400 fill-yellow-400" />
        Favourites
      </h2>
      <ul className="space-y-2">
        {favourites.map(r => (
          <li key={r.id}>
            <Link
              to={`/recipe/view/${r.id}`}
              className="flex items-center justify-between gap-3 px-3 py-2 rounded-lg hover:bg-gray-50 transition group"
            >
              <span className="text-sm font-medium text-gray-800 group-hover:text-indigo-600 truncate">
                {r.name}
              </span>
              <FiChevronRight size={14} className="text-gray-300 group-hover:text-indigo-400 shrink-0 transition" />
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}

function RecentRecipes({ recipes }) {
  if (!recipes || recipes.length === 0) return null;

  const recent = recipes.slice(-5).reverse();

  return (
    <div className="bg-white rounded-xl border border-gray-200 shadow-sm p-6">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold text-gray-900 flex items-center gap-2">
          <FiBook size={18} className="text-indigo-500" />
          Recent Recipes
        </h2>
        <Link to="/recipe" className="text-sm text-indigo-600 hover:text-indigo-800 font-medium transition">
          View all →
        </Link>
      </div>
      <ul className="space-y-2">
        {recent.map(r => (
          <li key={r.id}>
            <Link
              to={`/recipe/view/${r.id}`}
              className="flex items-center justify-between gap-3 px-3 py-2 rounded-lg hover:bg-gray-50 transition group"
            >
              <span className="text-sm font-medium text-gray-800 group-hover:text-indigo-600 truncate">
                {r.name}
              </span>
              {r.is_favourite && (
                <FiStar size={13} className="text-yellow-400 fill-yellow-400 shrink-0" />
              )}
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}

export function HomePage() {
  const { session } = useAuth();

  const { data: recipes } = useQuery({ queryKey: ['recipes'], queryFn: getRecipes });
  const { data: mealPlans } = useQuery({ queryKey: ['meal-plans'], queryFn: getMealPlans });
  const { data: groceryLists } = useQuery({ queryKey: ['grocery-lists'], queryFn: getGroceryLists });

  const firstName = session?.name?.split(' ')[0] ?? session?.name;

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">
            Welcome back{firstName ? `, ${firstName}` : ''}
          </h1>
          <p className="text-gray-500 mt-1">What are we cooking today?</p>
        </div>
        <QuickActions />
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <StatCard
          icon={FiBook}
          label="Recipes"
          count={recipes?.length}
          to="/recipe"
          color="bg-indigo-500"
        />
        <StatCard
          icon={FiCalendar}
          label="Meal Plans"
          count={mealPlans?.length}
          to="/mealplan"
          color="bg-pink-500"
        />
        <StatCard
          icon={FiShoppingCart}
          label="Grocery Lists"
          count={groceryLists?.length}
          to="/grocery"
          color="bg-emerald-500"
        />
      </div>

      {/* Favourites + Recent */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
        <FavouriteRecipes recipes={recipes} />
        <RecentRecipes recipes={recipes} />
      </div>
    </div>
  );
}
