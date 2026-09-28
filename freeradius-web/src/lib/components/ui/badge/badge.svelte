<script lang="ts">
	import { badgeVariants, type BadgeVariant } from './index.js';
	import { cn } from '$lib/utils.js';
	export let variant: BadgeVariant = 'default';
	let klass: string | undefined = undefined;
	export { klass as class };
</script>

<span class={cn(badgeVariants({ variant }), klass)} {...$$restProps}>
	<slot />
</span>
